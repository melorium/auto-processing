package queue

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/avian-digital-forensics/auto-processing/generate/ruby"
	api "github.com/avian-digital-forensics/auto-processing/pkg/avian-api"
	"github.com/avian-digital-forensics/auto-processing/pkg/powershell"

	"github.com/jinzhu/gorm"
	ps "github.com/simonjanss/go-powershell"
)

const (
	sleepMinutes = 2
)

type Queue struct {
	db    *gorm.DB
	shell ps.Shell
	uri   string
}

// New returns a new queue
func New(db *gorm.DB, shell ps.Shell, uri string) Queue {
	return Queue{db: db, shell: shell, uri: uri}
}

func (q *Queue) Start() {
	log.Println("queue started")
	for {
		q.loop()
		time.Sleep(time.Duration(sleepMinutes * time.Minute))
	}
}

func (q *Queue) loop() {
	log.Printf("getting runners from queue")
	runners, err := getRunners(q.db)
	if err != nil {
		log.Printf("cannot get runners: %v", err)
		return
	}
	log.Printf("found %d runners from queue", len(runners))

	for _, runner := range runners {
		// check if the runners server is active
		var server api.Server
		query := q.db.Where("active = ? and hostname = ?", false, runner.Hostname)
		if query.First(&server).RecordNotFound() {
			log.Printf("server %s is already active", runner.Hostname)
			continue
		}

		log.Println(server)

		// Check to see if licence is active
		nms, err := activeLicence(q.db, runner.Nms, runner.Licence, runner.Workers)
		if err != nil {
			log.Printf("unable to get licence : %v", err)
			continue
		}

		// create a new run
		log.Printf("Creating new run: %s", runner.Name)
		run := q.newRun(runner, &server, nms)
		if err := run.setActive(); err != nil {
			log.Printf("setActive: %v", err)
			continue
		}

		go run.handle(run.start())
	}
}

type run struct {
	queue  *Queue
	runner *api.Runner
	server *api.Server
	nms    *api.Nms
}

func (q *Queue) newRun(runner *api.Runner, server *api.Server, nms *api.Nms) *run {
	return &run{
		queue:  q,
		runner: runner,
		server: server,
		nms:    nms,
	}
}

func (r *run) setActive() error {
	db := r.queue.db

	// Set runner to active and save to db
	if err := db.Model(&api.Runner{}).Where("id = ?", r.runner.ID).Update("active", true).Error; err != nil {
		return fmt.Errorf("Failed to set runner to active: %v", err)
	}

	// Set server to active and save to db
	if err := db.Model(&api.Server{}).Where("id = ?", r.server.ID).Update("active", true).Error; err != nil {
		return fmt.Errorf("Failed to set server to active: %v", err)
	}

	// Set new values to NMS
	r.nms.InUse = r.runner.Workers
	for _, lic := range r.nms.Licences {
		if lic.Type == r.runner.Licence {
			lic.InUse += 1
			if err := db.Save(&lic).Error; err != nil {
				return fmt.Errorf("Failed to update licence: %s %s : %v", r.nms.Address, lic.Type, err)
			}
		}
	}

	// Save NMS to db
	if err := db.Save(&r.nms).Error; err != nil {
		return fmt.Errorf("Failed to set nms to active: %v", err)
	}

	return nil
}

func (r *run) start() error {
	// Generate the ruby-script for the runner
	log.Printf("generating script for runner: %s", r.runner.Name)
	script, err := ruby.Generate(r.queue.uri, *r.runner)
	if err != nil {
		return fmt.Errorf("failed to generate script for runner: %s - %v", r.runner.Name, err)
	}

	// Start a new remote ps-session
	log.Printf("creating new ps-session @ %s", r.runner.Hostname)
	var client *powershell.Client
	if len(r.server.Username) == 0 {
		// Start NewClient without credentials if username is not in the server
		client, err = powershell.NewClient(r.runner.Hostname, r.queue.shell)
	} else {
		client, err = powershell.NewClientWithCredentials(
			r.runner.Hostname,
			r.queue.shell,
			r.server.Username,
			r.server.Password,
		)
	}

	if err != nil {
		return fmt.Errorf("failed to create remote-client for powershell: %v", err)
	}

	// Set nuix username as an env-variable
	if err := client.SetEnv("NUIX_USERNAME", r.nms.Username); err != nil {
		return fmt.Errorf("unable to set NUIX_USERNAME env-variable: %v", err)
	}

	// Set nuix password as an env-variable
	if err := client.SetEnv("NUIX_PASSWORD", r.nms.Password); err != nil {
		return fmt.Errorf("unable to set NUIX_PASSWORD env-variable: %v", err)
	}

	nuixPath := powershell.FormatPath(r.server.NuixPath)
	log.Printf("Formatted path from: %s to %s", r.server.NuixPath, nuixPath)

	scriptName := powershell.FormatFilename(r.runner.Name + ".gen.rb")
	log.Printf("creating script: %s to path: %s @ %s", scriptName, nuixPath, r.runner.Hostname)
	if err := client.CreateFile(nuixPath, scriptName, []byte(script)); err != nil {
		return fmt.Errorf("Failed to create script-file: %v", err)
	}
	defer client.RemoveFile(nuixPath, scriptName)

	log.Printf("Starting runner: %s - host: %s - licenceserver: %s - licencetype: %s",
		r.runner.Name, r.server.Hostname, r.nms.Address, r.runner.Licence)
	return client.RunWithCmd(
		nuixPath,
		"nuix_console.exe",
		"-Xmx"+r.runner.Xmx,
		"-Dnuix.registry.servers="+r.nms.Address,
		"-licencesourcetype", "server",
		"-licencesourcelocation", fmt.Sprintf("%s:%d", r.nms.Address, r.nms.Port),
		"-licencetype", r.runner.Licence,
		"-licenceworkers", fmt.Sprintf("%d", r.runner.Workers),
		"-signout",
		"-release",
		scriptName,
	)
}

func (r *run) handle(err error) {
	db := r.queue.db
	defer r.close()

	// get the latest runner info
	var runner api.Runner
	if err := db.First(&runner, r.runner.ID).Error; err != nil {
		log.Printf("failed to get runner-information for: %s : %v", r.runner.Name, err)
	}

	runner.Active = false
	runner.Finished = true

	// handle the error
	if err != nil {
		runner.Finished = false
		log.Printf("runner: %s failed : %v", r.runner.Name, err)
	}

	// update the runner to db
	if err := db.Save(&runner).Error; err != nil {
		log.Printf("Cannot save runner-information for %s : %v", r.runner.Name, err)
	}

	// Set server to inactive
	if err := db.Model(&api.Server{}).Where("id = ?", r.server.ID).Update("active", false).Error; err != nil {
		log.Printf("Failed to set server %s to in-active: %v", r.server.Hostname, err)
	}

	// Get the latest data for the nms-server
	var nms api.Nms
	if err := db.Preload("Licences").First(&nms, r.nms.ID).Error; err != nil {
		log.Printf("Cannot get nms %s from db : %v", r.nms.Address, err)
		return
	}

	// Reset the licences for the nms
	nms.InUse = nms.InUse - r.runner.Workers
	for _, lic := range nms.Licences {
		if lic.Type == r.runner.Licence {
			lic.InUse = lic.InUse - 1
			if err := db.Save(&lic).Error; err != nil {
				log.Printf("failed to reset licence for runner: %s", r.runner.Name)
				return
			}
		}
	}

	// update the nms to the db
	if err := db.Save(&nms).Error; err != nil {
		log.Printf("Failed to reset nms %s : %v", r.nms.Address, err)
	}

	return
}

func (r *run) close() error {
	if r == nil {
		return errors.New("run is already closed")
	}
	log.Printf("Closing runner: %v", r)

	r.queue = nil
	r.runner = nil
	r.server = nil
	r.nms = nil
	return nil
}

func getRunners(db *gorm.DB) ([]*api.Runner, error) {
	var runners []*api.Runner
	err := db.Preload("Stages.Process.EvidenceStore").
		Preload("Stages.SearchAndTag.Files").
		Preload("Stages.Exclude").
		Preload("Stages.Ocr").
		Preload("Stages.Reload").
		Preload("Stages.Populate.Types").
		Preload("CaseSettings.Case").
		Preload("CaseSettings.CompoundCase").
		Preload("CaseSettings.ReviewCompound").
		Where("active = ? and finished = ?", false, false).
		Find(&runners).Error
	return runners, err
}

func activeLicence(db *gorm.DB, address, licencetype string, workers int64) (*api.Nms, error) {
	// Get the requested NMS
	var nms api.Nms
	if err := db.Preload("Licences").First(&nms, "address = ?", address).Error; err != nil {
		return nil, err
	}

	// Check if we have available workers
	if workers > (nms.Workers - nms.InUse) {
		return nil, fmt.Errorf("not enough workers available - requested: %d - available: %d/%d", workers, nms.InUse, nms.Workers)
	}

	// Check if we have a free licence
	for _, lic := range nms.Licences {
		if lic.Type == licencetype {
			if lic.InUse < lic.Amount {
				return &nms, nil
			} else {
				return nil, fmt.Errorf("not enough licences available for %s - %d/%d in use", licencetype, lic.InUse, lic.Amount)
			}
		}
	}
	return nil, fmt.Errorf("did not find licencetype: %s", licencetype)
}
