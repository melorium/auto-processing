package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/avian-digital-forensics/auto-processing/config"
)

func main() {
	flags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	cfgPath := flags.String("cfg", "./configs/auto-processing.yml", "filepath for the config")
	if err := flags.Parse(os.Args[1:]); err != nil {
		fmt.Printf("Failed to get cfg, use flag: --cfg=/path/to/cfg - error: %v", err)
		os.Exit(2)
	}

	cfg, err := config.GetConfig(*cfgPath)
	if err != nil {
		fmt.Printf("Failed to get config: %s - %v", *cfgPath, err)
		os.Exit(2)
	}

	licenseCfg, err := config.GetConfig("./configs/license.yml")
	if err != nil {
		fmt.Printf("Failed to get config for license: %s - %v", *cfgPath, err)
		os.Exit(2)
	}

	cfg.Server = licenseCfg.Server

	if err := cfg.Validate(); err != nil {
		fmt.Printf("Error when validating config: %s - %v", *cfgPath, err)
		os.Exit(2)
	}
	fmt.Printf("Config OK : %s", *cfgPath)
}
