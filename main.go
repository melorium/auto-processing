package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/avian-digital-forensics/auto-processing/config"
)


func main() {
	dt := time.Now()
	timestamp := dt.Format("20060102")
	filename := fmt.Sprintf("%s%s.log", "auto-processing", timestamp)

	// create log-dir
	if _, err := os.Stat("./logs"); os.IsNotExist(err) {
		os.Mkdir("./logs", os.ModePerm)
	}

	// create log-file with read-write & create-permissions
	logFile, err := os.OpenFile("./logs/"+filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Failed to open logfile: %v", err)
		os.Exit(2)
	}
	defer logFile.Close()

	logger := &logrus.Logger{
		Out:   io.MultiWriter(os.Stdout, logFile),
		Level: logrus.DebugLevel,
		Formatter: new(logrus.JSONFormatter),
	}
	log := logger.WithField("source", "main.go")
	
	log.Info("Initializing script")

	cfg := config.GetConfig(".\\config.yml")
	log.Info("Validating config")
	if err := cfg.Validate(); err != nil {
		log.Error(err)
		os.Exit(2)
	}

	tmpDir := os.TempDir()
	defer os.RemoveAll(tmpDir)

	path, err := os.Getwd()
	if err != nil {
		log.Error("cant get working dir:", err)
		os.Exit(2)
	}

	cfg.Nuix.Settings.WorkingPath = path
	
	file, err := json.MarshalIndent(cfg.Nuix.Settings, "", " ")
	if err != nil {
		log.Error(err)
		os.Exit(2)
	}

	tmpFile := tmpDir + "/settings.json"
	if err = ioutil.WriteFile(tmpFile, file, 0644); err != nil {
		log.Error(err)
		os.Exit(2)
	}
	
	program := cfg.Server.NuixPath + "\\nuix_console.exe"

	cmd := exec.Command(
		program,
		"-Xmx" + cfg.Nuix.Xmx,
		"-Dnuix.registry.servers=" + cfg.Server.NmsAddress,
		"-licencesourcetype",
		"server",
		"-licencesourcelocation",
		cfg.Server.NmsAddress + ":27443",
		"-licencetype",
		cfg.Server.Licencetype,
		"-licenceworkers",
		cfg.Nuix.Workers,
		"-signout",
		path + "\\process.rb",
		"-s",
		tmpFile,
		strings.Join(cfg.Nuix.Switches , " "),
	)
	
	cmd.Dir = cfg.Server.NuixPath
	cmd.Env = append(os.Environ(),
		"NUIX_USERNAME=" + cfg.Server.Username,
		"NUIX_PASSWORD=" + cfg.Server.Password,
	)

	cmd.Stdout = io.MultiWriter(os.Stdout, logFile)
	cmd.Stderr = os.Stderr
	
	log.Info("Executing Nuix with command: ", cmd)

	if err := cmd.Start(); err != nil {
		log.Error("Failed to run program", err)	
		os.Exit(2)
	}
	if err := cmd.Wait(); err != nil {
		log.Error("Failed to run program", err)	
		os.Exit(2)
	}
}