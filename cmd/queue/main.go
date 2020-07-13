package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/avian-digital-forensics/auto-processing/config"
	"github.com/avian-digital-forensics/auto-processing/log"
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
	pwd     string
}

func main() {
	flags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	cfgPath := flags.String("cfg", "./configs/queue.yml", "filepath for the config")
	if err := flags.Parse(os.Args[1:]); err != nil {
		fmt.Printf("Failed to get cfg, use flag: --cfg=/path/to/cfg - error: %v", err)
		os.Exit(2)
	}

	pwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Failed to get pwd - error: %v", err)
		os.Exit(2)
	}

	log, logFile := log.Get("queue", *cfgPath)
	defer logFile.Close()

	app := app{log, *cfgPath, pwd}

	app.queue()

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

func (a *app) queue() {
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
		a.loopQueue(cfg)
		a.log.Debug("Going back to sleep")

		time.Sleep(time.Duration(sleepMinutes * time.Minute))
	}
}

func serverActive(host string, cfg *config.Config) bool {
	for _, queue := range cfg.Queue {
		if host == queue.Host {
			if queue.Active {
				return false
			}
		}
	}
	return true
}

func (a *app) runRemote(queue *config.Queue, cfg *config.Config) {
	cfgPath := fmt.Sprintf("%s/%s", queue.ProgramPath, queue.Config)

	cmd := exec.Command("powershell.exe",
		fmt.Sprintf("%s\\scripts\\powershell\\RunRemote.ps1", a.pwd),
		fmt.Sprintf("-ComputerName %s", queue.Host),
		fmt.Sprintf("-ProgramPath %s", queue.ProgramPath),
		fmt.Sprintf("-Config %s\\%s\\%s", a.pwd, "configs", queue.Config),
		fmt.Sprintf("-Destination %s", cfgPath),
	)

	a.handleCMD(cmd, queue, cfg)
}

func (a *app) handleCMD(cmd *exec.Cmd, queue *config.Queue, cfg *config.Config) {
	a.log.Infof("Running command: %s", cmd.String())
	if err := cmd.Start(); err != nil {
		a.log.Errorf("Failed to run program with config: %s - error: %v", queue.Config, err)
		queue.SetFailed()
		a.updateQueue(cfg)
		return
	}
	if err := cmd.Wait(); err != nil {
		a.log.Errorf("Failed to run program with config: %s - error: %v", queue.Config, err)
		queue.SetFailed()
		a.updateQueue(cfg)
		return
	}

	queue.SetSuccessful()
	a.updateQueue(cfg)
	a.log.Info("Processing finished for config: ", queue.Config)
}

func (a *app) runLocal(queue *config.Queue, cfg *config.Config) {
	a.log.Info("Running program locally")
	cmd := exec.Command("./auto-processing.exe", "--cfg=./configs/"+queue.Config)
	a.handleCMD(cmd, queue, cfg)
}

func (a *app) loopQueue(cfg *config.Config) {
	for _, queue := range cfg.Queue {
		if queue.Successful {
			a.log.Debugf("Skipping the run with config: %s since it already has been processed", queue.Config)
			continue
		}
		if queue.Active {
			a.log.Debugf("Skipping run with config: %s queue is already active for this config", queue.Config)
			continue
		}

		a.log.Infoln("Starting auto-processing with config:", queue.Config)
		queue.SetActive()
		a.updateQueue(cfg)

		if queue.Host != "" && queue.ProgramPath != "" {
			if serverActive(queue.Host, cfg) {
				a.log.Warnf("Host %s is already active, skipping", queue.Host)
				return
			}

			a.log.Infof("Starting remote-process @ %s", queue.Host)
			go a.runRemote(queue, cfg)

		} else {
			a.runLocal(queue, cfg)
		}
	}
}
