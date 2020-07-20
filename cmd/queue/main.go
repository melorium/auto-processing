package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/avian-digital-forensics/auto-processing/cmd/queue/app"
	"github.com/avian-digital-forensics/auto-processing/log"
)

func main() {
	flags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	cfgPath := flags.String("cfg", "queue.yml", "filepath for the config")
	if err := flags.Parse(os.Args[1:]); err != nil {
		fmt.Printf("Failed to get cfg, use flag: --cfg=/path/to/cfg - error: %v", err)
		os.Exit(2)
	}

	pwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Failed to get pwd - error: %v", err)
		os.Exit(2)
	}

	tmpDir, err := ioutil.TempDir(pwd, "tmp")
	if err != nil {
		fmt.Printf("Failed to create temp-directory - error: %v", err)
		os.Exit(2)
	}
	defer os.RemoveAll(tmpDir)

	log, logFile := log.Get("queue", *cfgPath)
	defer logFile.Close()

	app := app.App{
		Log:     log,
		CfgPath: *cfgPath,
		Pwd:     pwd,
		TmpDir:  tmpDir,
	}

	app.Start()
}
