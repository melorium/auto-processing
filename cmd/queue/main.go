package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"

	"github.com/avian-digital-forensics/auto-processing/cmd/queue/app"
	"github.com/avian-digital-forensics/auto-processing/log"
)

func main() {
	if err := run(os.Args, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(2)
	}
}

func run(args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	cfgPath := flags.String("cfg", "queue.yml", "filepath for the config")
	if err := flags.Parse(os.Args[1:]); err != nil {
		return fmt.Errorf("Failed to get cfg, use flag: --cfg=/path/to/cfg - error: %v", err)
	}

	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Failed to get pwd - error: %v", err)
	}

	tmpDir, err := ioutil.TempDir(pwd, "tmp")
	if err != nil {
		return fmt.Errorf("Failed to create temp-directory - error: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	log, logFile := log.Get("queue", *cfgPath)
	defer logFile.Close()

	app := app.App{
		Log:     log,
		LogFile: logFile,
		CfgPath: *cfgPath,
		Pwd:     pwd,
		TmpDir:  tmpDir,
	}

	// Listen to os-signal and remove tmpdir
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			fmt.Fprintf(stdout, "signal: %v", sig)
			if err := os.RemoveAll(tmpDir); err != nil {
				log.Errorf("Cannot remove temp-dir: %v", err)
			}
			os.Exit(2)
		}
	}()

	app.Start()
	return nil
}
