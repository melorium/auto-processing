package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/avian-digital-forensics/auto-processing/config"
)

func main() {
	log.Println("Initializing script")

	cfg := config.GetConfig(".\\config.yml")
	if err := cfg.Validate(); err != nil {
		log.Println(err)
		os.Exit(2)
	}

	tmpDir := os.TempDir()
	defer os.RemoveAll(tmpDir)

	path, err := os.Getwd()
	if err != nil {
		log.Println("cant get working dir:", err)
		os.Exit(2)
	}

	cfg.Nuix.Settings.WorkingPath = path
	
	file, err := json.MarshalIndent(cfg.Nuix.Settings, "", " ")
	if err != nil {
		log.Println(err)
		os.Exit(2)
	}

	tmpFile := tmpDir + "/settings.json"
	if err = ioutil.WriteFile(tmpFile, file, 0644); err != nil {
		log.Println(err)
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
		path + "\\process.rb",
		"-p",
		cfg.Nuix.ProcessProfilePath,
		"-n",
		cfg.Nuix.ProcessProfileName,
		"-s",
		tmpFile,
		strings.Join(cfg.Nuix.Switches , " "),
	)
	
	cmd.Dir = cfg.Server.NuixPath
	cmd.Env = append(os.Environ(),
		"NUIX_USERNAME=" + cfg.Server.Username,
		"NUIX_PASSWORD=" + cfg.Server.Password,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	log.Println("Executing Nuix with command: ", cmd)

	if err := cmd.Start(); err != nil {
		log.Println("Failed to run program", err)	
		os.Exit(2)
	}
	if err := cmd.Wait(); err != nil {
		log.Println("Failed to run program", err)	
		os.Exit(2)
	}
}