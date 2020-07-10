package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/avian-digital-forensics/auto-processing/config"
	"github.com/avian-digital-forensics/auto-processing/log"
	"github.com/avian-digital-forensics/auto-processing/pkg/powershell"
	ps "github.com/simonjanss/go-powershell"
	"github.com/simonjanss/go-powershell/backend"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

const (
	sleepMinutes       = 5
	sleepMinutesFailed = 2
)

type app struct {
	log     *logrus.Entry
	cfgPath string
}

func main() {
	flags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	cfgPath := flags.String("cfg", "./configs/queue.yml", "filepath for the config")
	if err := flags.Parse(os.Args[1:]); err != nil {
		fmt.Printf("Failed to get cfg, use flag: --cfg=/path/to/cfg - error: %v", err)
		os.Exit(2)
	}

	log, logFile := log.Get("queue", *cfgPath)
	defer logFile.Close()

	app := app{log, *cfgPath}

	// start a local powershell process
	shell, err := ps.New(&backend.Local{})
	if err != nil {
		fmt.Printf("Failed to start powershell process - error: %v", err)
		os.Exit(2)
	}
	defer shell.Close()

	app.queue(shell)

}

func (a *app) updateQueue(cfg *config.Config) {
	file, err := os.OpenFile(a.cfgPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		a.log.Error("failed to update queue:", err)
		return
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		a.log.Error("failed to update queue:", err)
		return
	}

	_, err = file.Write(data)
	if err != nil {
		a.log.Error("failed to update queue:", err)
		return
	}
	return
}

func (a *app) queue(shell ps.Shell) {
	a.log.Info("Starting queue")
	for {
		cfg, err := config.GetConfig(a.cfgPath)
		if err != nil {
			a.log.Fatalf("Could not open the config - %v", err)
			os.Exit(2)
		}
		if len(cfg.Queue) < 1 {
			a.log.Debug("Queue is empty - go back to sleep")
			continue
		}

		a.log.Debug("Looking for queue")
		a.loopQueue(cfg, shell)
		a.log.Debug("Going back to sleep")
		time.Sleep(time.Duration(sleepMinutes * time.Minute))
	}
}

func serverActive(host string, cfg *config.Config) bool {
	for _, queue := range cfg.Queue {
		if host == *queue.Host {
			if queue.Active {
				return false
			}
		}
	}
	return true
}

func (a *app) runRemote(queue *config.Queue, cfg *config.Config, sh ps.Shell) {
	client, err := powershell.NewClient(*queue.Host, *queue.ProgramPath, sh)
	if err != nil {
		a.log.Errorf("Failed to create remote-client: %v", err)
		queue.SetFailed()
		a.updateQueue(cfg)
		return
	}
	defer client.Close()

	if err := client.Run("config.yml"); err != nil {
		a.log.Error("Failed to run remote-job")
		queue.SetFailed()
		a.updateQueue(cfg)
		return
	}

	queue.SetSuccessful()
	a.updateQueue(cfg)
	a.log.Info("Processing finished for config: ", queue.Config)
}

func (a *app) runLocal(queue *config.Queue, cfg *config.Config) {
	cmd := exec.Command("./auto-processing.exe", "--cfg="+queue.Config)
	if err := cmd.Start(); err != nil {
		a.log.Error("Failed to run program", err)
		queue.SetFailed()
		a.updateQueue(cfg)
		return
	}
	if err := cmd.Wait(); err != nil {
		a.log.Error("Failed to run program", err)
		queue.SetFailed()
		a.updateQueue(cfg)
		return
	}

	queue.SetSuccessful()
	a.updateQueue(cfg)
	a.log.Info("Processing finished for config: ", queue.Config)
}

func (a *app) loopQueue(cfg *config.Config, sh ps.Shell) {
	start := true
	for _, queue := range cfg.Queue {
		if queue.Successful {
			a.log.Debugf("Skipping the run with config: %s since it already has been processed", queue.Config)
			continue
		}
		if queue.Active {
			a.log.Error("Something is wrong, queue is already active for this config:", queue.Config)
			continue
		}

		if !start {
			a.log.Debugf("Sleeping for %v minutes to wait for license to be released", sleepMinutesFailed)
			time.Sleep(time.Duration(sleepMinutesFailed * time.Minute))
		}
		start = false

		a.log.Infoln("Starting auto-processing with config:", queue.Config)
		queue.SetActive()
		a.updateQueue(cfg)

		if queue.Host != nil && queue.ProgramPath != nil {
			if serverActive(*queue.Host, cfg) {
				a.log.Warnf("Host %s is already active, skipping", *queue.Host)
				return
			}

			go a.runRemote(queue, cfg, sh)

		} else {
			a.runLocal(queue, cfg)
		}
	}
}
