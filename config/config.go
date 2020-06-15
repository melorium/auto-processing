package config

import (
	"fmt"
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
	CaseLocation string `yaml:"case_location" json:"case_location"`
	Profile string `yaml:"profile" json:"profile"`
	ProfileLocation string `yaml:"profile_location" json:"profile_location"`
	Compound bool `yaml:"compound" json:"compound"`
	CompoundCase *Case `yaml:"compound_case" json:"compound_case"`
	ReviewCompound *Case `yaml:"review_compound" json:"review_compound"`
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
	Case Case `yaml:"case" json:"case"`
	Reason string `yaml:"reason" json:"reason"`
	Files []string `yaml:"files" json:"files"`
}

func (cfg *Config) Validate() error {
	// Check if the nuix-path exists
	if ok, err := isReadable(cfg.Server.NuixPath); !ok && err != nil {
		return fmt.Errorf("No read access to NuixPath: %v", err)
	}
	
	// Check if the process-profile is readable
	if ok, err := isReadable(cfg.Nuix.Settings.ProfileLocation); !ok && err != nil {
		return fmt.Errorf("No write read to ProfileLocation: %v", err)
	}

	// Check if the master case location is writable
	if ok, err := isWritable(cfg.Nuix.Settings.CaseLocation); !ok && err != nil {
		return fmt.Errorf("No write access to CaseLocation: %v", err)
	}

	cfg.Nuix.Settings.CompoundCase.Directory = cfg.Nuix.Settings.CaseLocation + "/compound"

	// The user might need to configure the directory for review-compound
	if len(cfg.Nuix.Settings.ReviewCompound.Directory) == 0 {
		cfg.Nuix.Settings.ReviewCompound.Directory = fmt.Sprintf("%s/review-compound", cfg.Nuix.Settings.CaseLocation)
		cfg.Nuix.Settings.ReviewCompound.Name = "review-compound"
	} else {
		// Check that review-compound dir is writable
		if ok, err := isWritable(cfg.Nuix.Settings.ReviewCompound.Directory); !ok && err != nil {
			return fmt.Errorf("No write access to ReviewCompound.Directory: %v", err)
		}
	}
	
	// set a single-case name and directory based on directory availability
	collection := 1
	for {
		singleCaseDir := fmt.Sprintf("%s/single-c%d", cfg.Nuix.Settings.CaseLocation, collection)
		if _, err := os.Stat(singleCaseDir); os.IsNotExist(err) {
			cfg.Nuix.Settings.Case.Directory = singleCaseDir
			cfg.Nuix.Settings.Case.Name = fmt.Sprintf("single-c%d", collection)
			break
		}
		collection++
	}

	// Check if the evidences are readable
	for _, evidence := range cfg.Nuix.Settings.EvidenceStore {
		if ok, err := isReadable(evidence.Directory); !ok && err != nil {
			return err
		}
	}

	// Check the configuration for the sub-steps
	for _, subStep := range cfg.Nuix.Settings.SubSteps {
		// Check if the profile location if the user has provided one
		if len(subStep.ProfileLocation) != 0 {
			if ok, err := isReadable(subStep.ProfileLocation); !ok && err != nil {
				return err
			}
		}

		// Make sure to have right-access to the sub-case directory
		if len(subStep.Case.Directory) != 0 {
			if ok, err := isWritable(subStep.Case.Directory); !ok && err != nil {
				return err
			}
			break
		}
	
		// Set automatic directory and name for the subcase (since it has not been configured by the user)
		review := 1
		for {
			reviewCaseDir := fmt.Sprintf("%s/review-c%d-r%d", cfg.Nuix.Settings.CaseLocation, collection, review)
			if _, err := os.Stat(reviewCaseDir); os.IsNotExist(err) {
				subStep.Case.Directory = reviewCaseDir
				subStep.Case.Name = fmt.Sprintf("review-c%d-r%d", collection, review)
				break
			}
			review++
		}
	}

	// Function to check write-access for directory from the switches
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