package app

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	"github.com/avian-digital-forensics/auto-processing/config"
	"github.com/avian-digital-forensics/auto-processing/log"
	"github.com/avian-digital-forensics/auto-processing/pkg/powershell"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

const (
	sleepMinutes       = 5
	sleepMinutesFailed = 2
)

type App struct {
	Log     *logrus.Entry
	LogFile *os.File
	CfgPath string
	Pwd     string
	TmpDir  string
}

func (a *App) Start() {
	a.Log.Info("Starting queue")
	for {
		cfg, err := config.GetConfig(a.CfgPath)
		if err != nil {
			a.Log.Errorf("Could not open the config - %v", err)
			os.Exit(2)
		}
		if len(cfg.Queue) < 1 {
			a.Log.Debug("Queue is empty - go back to sleep")
			continue
		}

		a.Log.Debug("Looking for queue")
		a.loop(cfg)
		a.Log.Debug("Going back to sleep")

		time.Sleep(time.Duration(sleepMinutes * time.Minute))
	}
}

func (a *App) loop(cfg *config.Config) {
	a.Log.Warnf("Queue: %v", cfg.Queue)
	for _, runner := range cfg.Queue {
		a.Log.Debugf("Config: %s Host: %s", runner.Config, runner.Host)
		// Check if the runner already has run
		if runner.Successful {
			a.Log.Debugf("Skipping the run with config: %s since it already has been processed", runner.Config)
			continue
		}

		// Check if the runner is active
		if runner.Active {
			a.Log.Debugf("Skipping runner with config: %s queue is already active for this config", runner.Config)
			continue
		}

		// Check if the server is active
		if serverActive(runner.Host, cfg) {
			a.Log.Debugf("Host %s is already active, skipping", runner.Host)
			continue
		}

		runner.SetActive()
		a.updateQueue(cfg)
		if runner.Host != "localhost" {
			a.Log.Infof("Starting remote-process @ %s", runner.Host)
			go a.runRemote(runner, cfg)
		} else {
			a.Log.Infof("Starting local-process @ %s", runner.Host)
			go a.runLocal(runner, cfg)
		}
	}
}

func (a *App) updateQueue(cfg *config.Config) {
	file, err := os.OpenFile("./configs/"+a.CfgPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		a.Log.Error("failed to update queue:", err)
		return
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		a.Log.Error("failed to update queue:", err)
		return
	}

	_, err = file.Write(data)
	if err != nil {
		a.Log.Error("failed to update queue:", err)
		return
	}
	return
}

func serverActive(host string, cfg *config.Config) bool {
	for _, queue := range cfg.Queue {
		if host == queue.Host {
			if queue.Active {
				return true
			}
		}
	}
	return false
}

func (a *App) getFiles(dir string) []string {
	dirs, err := ioutil.ReadDir(dir)
	if err != nil {
		a.Log.Errorf("Cannot read directory %v", err)
		return nil
	}

	var files []string
	for _, file := range dirs {
		if !file.IsDir() {
			files = append(files, dir+file.Name())
		} else if file.IsDir() {
			// Recurse
			newDir := dir + file.Name() + "/"
			files = append(files, a.getFiles(newDir)...)
		}
	}
	return files
}

func (a *App) runRemote(runner *config.Queue, cfg *config.Config) {
	var files []string
	files = a.getFiles("./scripts/ruby/")
	files = append(files, "auto-processing.exe", "./configs/"+runner.Config, "./configs/license.yml")

	// Archive the scripts, config and the executable binary
	archive, err := a.archiveFiles(files)
	if err != nil {
		a.Log.Errorf("Could not archive files: %v", err)
		runner.SetFailed()
		a.updateQueue(cfg)
		return
	}
	defer os.Remove(archive.Path)

	_, srvLog := log.Get(runner.Host, "")
	defer srvLog.Close()

	// Create a ps-client to use for the remote-connection
	ps := powershell.NewClient(runner.Host, runner.Username, runner.Password, srvLog)
	if err := ps.AutoProcessing(archive.Name, archive.Path, runner.Config); err != nil {
		a.Log.Errorf("Failed to run program @ %s : %v", runner.Host, err)
		runner.SetFailed()
		a.updateQueue(cfg)
		return
	}
}

func (a *App) handleCMD(cmd *exec.Cmd, queue *config.Queue, cfg *config.Config) {
	a.Log.Infof("Running command: %s", cmd.String())
	if err := cmd.Start(); err != nil {
		a.Log.Errorf("Failed to run program with config: %s - error: %v", queue.Config, err)
		queue.SetFailed()
		a.updateQueue(cfg)
		return
	}
	if err := cmd.Wait(); err != nil {
		a.Log.Errorf("Failed to run program with config: %s - error: %v", queue.Config, err)
		queue.SetFailed()
		a.updateQueue(cfg)
		return
	}

	queue.SetSuccessful()
	a.updateQueue(cfg)
	a.Log.Info("Processing finished for config: ", queue.Config)
}

func (a *App) runLocal(queue *config.Queue, cfg *config.Config) {
	a.Log.Info("Running program locally")
	cmd := exec.Command("./auto-processing.exe", "--cfg="+queue.Config)
	a.handleCMD(cmd, queue, cfg)
}

func appendFiles(filename string, writer *zip.Writer) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("Failed to open %s: %s", filename, err)
	}
	defer file.Close()

	wr, err := writer.Create(filename)
	if err != nil {
		return fmt.Errorf("Failed to create entry for %s in zip file: %v", filename, err)
	}

	if _, err := io.Copy(wr, file); err != nil {
		return fmt.Errorf("Failed to write %s to zip: %s", filename, err)
	}

	return nil
}

type zipFile struct {
	Name string
	Path string
}

func (a *App) archiveFiles(files []string) (*zipFile, error) {
	uuid, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	zipName := fmt.Sprintf("%v.zip", uuid)
	file, err := os.Create(fmt.Sprintf("%s\\%s", a.TmpDir, zipName))
	if err != nil {
		//os.Remove(file.Name())
		return nil, fmt.Errorf("Failed to create file for writing: %v", err)
	}
	defer file.Close()

	writer := zip.NewWriter(file)
	defer writer.Close()

	for _, filename := range files {
		if err := appendFiles(filename, writer); err != nil {
			return nil, fmt.Errorf("Failed to add file %s to zip: %v", filename, err)
		}
	}
	zipped := zipFile{Name: zipName, Path: file.Name()}
	return &zipped, nil
}
