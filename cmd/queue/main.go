package main

import (
	"flag"
	"os"
	"os/exec"
	"time"

	"github.com/avian-digital-forensics/auto-processing/config"
	"github.com/avian-digital-forensics/auto-processing/log"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

const (
	sleepMinutes = 5
	sleepMinutesFailed = 2
)

func main() {
	log, logFile := log.Get("queue")
	defer logFile.Close()

	flags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	cfgPath := flags.String("cfg", "./config.yml", "filepath for the config")
	if err := flags.Parse(os.Args[1:]); err != nil {
		log.Error(err)
		os.Exit(2)
	}

	queue(log, cfgPath)

}

func updateQueue(path string, cfg *config.Config) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func queue(log *logrus.Entry, cfgPath *string) {
	log.Info("Starting queue")
	for {
			cfg, err := config.GetConfig(*cfgPath)
			if err != nil {
				log.Fatalf("Could not open the config - %v", err)
				os.Exit(2)
			}
			if len(cfg.Queue) < 1 {
				log.Debug("Queue is empty - go back to sleep")
				continue
			}
		
			log.Debug("Looking for queue")
			loopQueue(log, cfg, cfgPath)
			log.Debug("Going back to sleep")
			time.Sleep(time.Duration(sleepMinutes * time.Minute))
	}
}

func loopQueue(log *logrus.Entry, cfg *config.Config, cfgPath *string) {
	start := true
	for _, queue := range cfg.Queue {
		if queue.Successful {
			log.Debugf("Skipping the run with config: %s since it already has been processed", queue.Config)
			continue
		}
		if queue.Active {
			log.Error("Something is wrong, queue is already active for this config:", queue.Config)
			continue
		}

		if !start {
			log.Debugf("Sleeping for %v minutes to wait for license to be released", sleepMinutesFailed)
			time.Sleep(time.Duration(sleepMinutesFailed * time.Minute))
		}
		start = false

		log.Infoln("Starting auto-processing with config:", queue.Config)
		queue.SetActive()
		if err := updateQueue(*cfgPath, cfg); err != nil {
			log.Error("failed to update queue:", err)
			continue
		}

		cmd := exec.Command("./auto-processing.exe", "--cfg=" + queue.Config)
		if err := cmd.Start(); err != nil {
			log.Error("Failed to run program", err)
			queue.SetFailed()
			if err := updateQueue(*cfgPath, cfg); err != nil {
				log.Error("failed to update queue: ", err)
				continue
			}
			continue
		}
		if err := cmd.Wait(); err != nil {
			log.Error("Failed to run program", err)
			queue.SetFailed()
			if err := updateQueue(*cfgPath, cfg); err != nil {
				log.Error("failed to update queue: ", err)
				continue
			}
			continue
		}

		queue.SetSuccessful()
		if err := updateQueue(*cfgPath, cfg); err != nil {
			log.Error("failed to update queue: ", err)
			continue
		}
		log.Info("Processing finished for config: ", queue.Config)
	}
}