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
	Queue []*Queue `yaml:"queue"`
}

type Queue struct {
	Config string `yaml:"config"`
	Active bool `yaml:"active"`
	Successful bool `yaml:"successful"`
	Failed bool `yaml:"failed"`
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
	Settings *Settings `yaml:"settings"`
	Switches []string `yaml:"switches"`
}

type Settings struct {
	ProcessProfilePath string `yaml:"process_profile_path" json:"process_profile_path"`
	ProcessProfileName string `yaml:"process_profile_name" json:"process_profile_name"`
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
	ExportPath string `yaml:"export_path" json:"export_path"`
	Compound *Case `yaml:"compound_case"`
	Reason string `yaml:"reason" json:"reason"`
	Files []string `yaml:"files" json:"files"`
}

func (cfg *Config) Validate() error {
	log.Println("Validating config")
	
	// Check if the nuix-path exists
	if ok, err := isReadable(cfg.Server.NuixPath); !ok && err != nil {
		return err
	}
	
	// Check if the process-profile is readable
	if ok, err := isReadable(cfg.Nuix.Settings.ProcessProfilePath); !ok && err != nil {
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
		if len(subStep.ExportPath) > 0 {
			if ok, err := isWritable(subStep.ExportPath); !ok && err != nil {
				return err
			}
		}
	}

	checkSwitch := func(nuixSwitch, formatted string) error {
		if nuixSwitch[0:len(formatted)] == formatted {
			if ok, err := isWritable(nuixSwitch[len(formatted):]); !ok && err != nil {
				return err
			}
		}
		return nil
	}

	for _, nuixSwitch := range cfg.Nuix.Switches {
		// Check if the shared temp dir is writable
		sharedTempDir := "-Dnuix.processing.sharedTempDirectory="
		if err := checkSwitch(nuixSwitch, sharedTempDir); err != nil {
			return err
		}
		// Check if the shared temp dir is writable
		workerTempDir := "-Dnuix.worker.tmpdir="
		if err := checkSwitch(nuixSwitch, workerTempDir); err != nil {
			return err
		}
		// Check if the shared temp dir is writable
		javaTempDir := "-Djava.io.tmpdir="
		if err := checkSwitch(nuixSwitch, javaTempDir); err != nil {
			return err
		}
		// Check if the shared temp dir is writable
		exportSpoolDir := "-Dnuix.export.spoolDir="
		if err := checkSwitch(nuixSwitch, exportSpoolDir); err != nil {
			return err
		}
		// Check if the log dir is writable
		logDir := "-Dnuix.logdir="
		if err := checkSwitch(nuixSwitch, logDir); err != nil {
			return err
		}
	}
	
	return nil
}

func (q *Queue) SetActive() {
	q.Active = true
	q.Failed = false
	q.Successful = false
}

func (q *Queue) SetSuccessful() {
	q.Successful = true
	q.Active = false
	q.Failed = false
}

func (q *Queue) SetFailed() {
	q.Failed = true
	q.Successful = false
	q.Active = false
}

func isWritable(path string) (bool, error) {
    info, err := os.Stat(path)
    if err != nil {
        if os.IsNotExist(err) {
			if err := os.MkdirAll(path, os.ModePerm); err != nil {
				return false, fmt.Errorf("failed to create path at: %s - %v", path, err)
			}
		} else {
			return false, fmt.Errorf("error checking filepath at: %s - %v", path, err)
		}
		info, err = os.Stat(path)
		if err != nil {
			return false, fmt.Errorf("error checking filepath at: %s", path)
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
			return false, fmt.Errorf("Unable to read from: %s - %v", path, err)
		}
		return false, fmt.Errorf("problem with checking the file permission for: %s - %v", path, err)
	}
	file.Close()
	return true, nil
}

func readYAML(path string, cfg *Config) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	decoder := yaml.NewDecoder(file)
	if err = decoder.Decode(cfg); err != nil {
		return err
	}
	return nil
}

// GetConfig returns data from config.yml
func GetConfig(path string) (*Config, error) {
	var cfg Config
	err := readYAML(path, &cfg)
	return &cfg, err
}