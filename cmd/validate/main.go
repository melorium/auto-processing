package main

import (
	"flag"
	"fmt"

	"github.com/avian-digital-forensics/auto-processing/config"
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return "my string representation"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var configs arrayFlags

func main() {
	flag.Var(&configs, "cfg", "Path for the config to validate")
	flag.Parse()

	var result []string

	for _, cfgPath := range configs {
		cfg, err := config.GetConfig(cfgPath)
		if err != nil {
			msg := fmt.Sprintf("Validate FAILED: %s - Failed to get config: %v", cfgPath, err)
			result = append(result, msg)
			continue
		}

		licenseCfg, err := config.GetConfig("./configs/license.yml")
		if err != nil {
			msg := fmt.Sprintf("Validate FAILED: %s - Failed to get license-config: %v", cfgPath, err)
			result = append(result, msg)
			continue
		}

		cfg.Server = licenseCfg.Server

		if err := cfg.Validate(); err != nil {
			msg := fmt.Sprintf("Validate FAILED: %s - %v", cfgPath, err)
			result = append(result, msg)
			continue
		}
	}

	for _, msg := range result {
		fmt.Println(msg)
	}
}
