package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/avian-digital-forensics/auto-processing/config"
	"github.com/avian-digital-forensics/auto-processing/log"
)

func main() {
	flags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	cfgPath := flags.String("cfg", "auto-processing.yml", "filepath for the config")
	if err := flags.Parse(os.Args[1:]); err != nil {
		fmt.Printf("Error trying to find config: %v", err)
		os.Exit(2)
	}

	log, logFile := log.Get("auto-processing", *cfgPath)
	defer logFile.Close()

	log.Info("Initializing script")

	cfg, err := config.GetConfig(*cfgPath)
	if err != nil {
		log.Errorf("Failed to get config: %s - %v", *cfgPath, err)
		os.Exit(2)
	}

	licenseCfg, err := config.GetConfig("license.yml")
	if err != nil {
		log.Errorf("Failed to get config for license: %s - %v", *cfgPath, err)
		os.Exit(2)
	}

	cfg.Server = licenseCfg.Server

	log.Info("Validating config")
	if err := cfg.Validate(); err != nil {
		log.Error(err)
		os.Exit(2)
	}

	tmpDir := os.TempDir()
	defer os.RemoveAll(tmpDir)

	path, err := os.Getwd()
	if err != nil {
		log.Error("Unable to get working dir: ", err)
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
		"-Xmx"+cfg.Nuix.Xmx,
		"-Dnuix.registry.servers="+cfg.Server.NmsAddress,
		"-licencesourcetype",
		"server",
		"-licencesourcelocation",
		cfg.Server.NmsAddress+":27443",
		"-licencetype",
		cfg.Server.Licencetype,
		"-licenceworkers",
		cfg.Nuix.Workers,
		"-signout",
		"-release",
		path+"\\scripts\\ruby\\process.rb",
		"-s",
		tmpFile,
		"-c",
		*cfgPath,
		strings.Join(cfg.Nuix.Switches, " "),
	)

	cmd.Dir = cfg.Server.NuixPath
	cmd.Env = append(os.Environ(),
		"NUIX_USERNAME="+cfg.Server.Username,
		"NUIX_PASSWORD="+cfg.Server.Password,
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
