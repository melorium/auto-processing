package config

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Server *ServerCfg `yaml:"server"`
	Nuix *NuixCfg `yaml:"nuix"`
}

type ServerCfg struct {
	NmsAddress string `yaml:"nms_address"`
	NuixPath string `yaml:"nuix_path"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Licencetype string `yaml:"licencetype"`
}

type NuixCfg struct {
	Xmx string `yaml:"xmx"`
	Workers string `yaml:"workers"`
	ProcessProfilePath string `yaml:"process_profile_path"`
	ProcessProfileName string `yaml:"process_profile_name"`
	Settings *Settings `yaml:"settings"`
	Switches []string `yaml:"switches"`
}

type Settings struct {
	Compound bool `yaml:"compound" json:"compound"`
	CompoundCase *Case `yaml:"compound_case" json:"compound_case"`
	Case *Case `yaml:"case" json:"case_settings"`
	EvidenceStore []*Evidence `yaml:"evidence_store" json:"evidence_settings"`
	SubSteps []*SubStep `yaml:"sub_steps" json:"sub_steps"`
	WorkingPath string `json:"working_path"`
}

type Case struct {
	Name string `yaml:"name" json:"name"`
	Directory string `yaml:"directory" json:"directory"`
	Description string `yaml:"description" json:"description"`
	Investigator string `yaml:"investigator" json:"investigator"`
}

type Evidence struct {
	Name string `yaml:"name" json:"name"`
	Directory string `yaml:"directory" json:"directory"`
	Description string `yaml:"description" json:"description"`
	Encoding string `yaml:"encoding" json:"encoding"`
	Custodian string `yaml:"custodian" json:"custodian"`
	Locale string `yaml:"locale" json:"locale"`
}

type SubStep struct {
	Type string `yaml:"type" json:"type"`
	Name string `yaml:"name" json:"name"`
	Profile string `yaml:"profile" json:"profile"`
	ProfileLocation string `yaml:"profile_location" json:"profile_location"`
	Search string `yaml:"search" json:"search"`
	Tag string `yaml:"tag" json:"tag"`
}

func (cfg *Config) Validate() error {
	log.Println("Validating config")
	
	// Check if the nuix-path exists
	if ok, err := isReadable(cfg.Server.NuixPath); !ok && err != nil {
		return err
	}
	
	// Check if the process-profile is readable
	if ok, err := isReadable(cfg.Nuix.ProcessProfilePath); !ok && err != nil {
		return err
	}

	// Check if the compound-case directory is writable
	if ok, err := isWritable(cfg.Nuix.Settings.CompoundCase.Directory); !ok && err != nil {
		return err
	}

	// Check if the case directory is writable
	if ok, err := isWritable(cfg.Nuix.Settings.Case.Directory); !ok && err != nil {
		return err
	}

	// Check if the evidences are readable
	for _, evidence := range cfg.Nuix.Settings.EvidenceStore {
		if ok, err := isReadable(evidence.Directory); !ok && err != nil {
			return err
		}
	}

	// Check if the profile is readable for the sub-steps
	for _, subStep := range cfg.Nuix.Settings.SubSteps {
		if len(subStep.ProfileLocation) > 0 {
			if ok, err := isReadable(subStep.ProfileLocation); !ok && err != nil {
				return err
			}
		}
	}
	
	return nil
}

func isWritable(path string) (bool, error) {
    info, err := os.Stat(path)
    if err != nil {
        if os.IsNotExist(err) {
			if err := os.MkdirAll(path, os.ModePerm); err != nil {
				log.Println("failed to create path at:", path)
				return false, err
			}
		} else {
			log.Println("error checking filepath at:", path)
			return false, err
		}
		info, err = os.Stat(path)
		if err != nil {
			log.Println("error checking filepath at:", path)
			return false, err
		}
    }

    if !info.IsDir() {
        return false, fmt.Errorf("path provided is not a directory: %s", path)
    }

    // Check if the user bit is enabled in file permission
    if info.Mode().Perm()&(1<<(uint(7))) == 0 {
        return false, fmt.Errorf("write permission bit is not set on this file for user: %s", path)
    }

    return true, nil
}

func isReadable(path string) (bool, error) {
	file, err := os.OpenFile(path, os.O_RDONLY, 0666)
	if err != nil {
		if os.IsPermission(err) {
			log.Println("Unable to read from ", path)
			return false, err
		}
		log.Println("problem with checking the file permission for:", path)
		return false, err
	}
	file.Close()
	return true, nil
}

func readYAML(path string, cfg *Config) {
	file, err := os.Open(path)
	if err != nil {
		log.Println(err)
		os.Exit(2)
	}

	decoder := yaml.NewDecoder(file)
	if err = decoder.Decode(cfg); err != nil {
		log.Println(err)
		os.Exit(2)
	}
}

// GetConfig returns data from config.yml
func GetConfig(path string) *Config {
	var cfg Config
	readYAML(path, &cfg)
	return &cfg
}