package configs

import (
	"errors"
	"fmt"
	"os"

	avian "github.com/avian-digital-forensics/auto-processing/pkg/avian-client"

	"gopkg.in/yaml.v2"
)

type Config struct {
	API API `yaml:"api"`
}

type API struct {
	Servers []Servers                `yaml:"servers"`
	Nms     avian.NmsApplyRequests   `yaml:"nmsApply"`
	Runner  avian.RunnerApplyRequest `yaml:"runner"`
}

type Servers struct {
	Server avian.ServerApplyRequest `yaml:"server"`
}

func SetCaseSettings(r avian.RunnerApplyRequest) (avian.RunnerApplyRequest, error) {
	var setCaseSettings bool
	for _, stage := range r.Stages {
		if stage.Process != nil {
			setCaseSettings = true
		}
	}

	if !setCaseSettings {
		r.CaseSettings = nil
		return r, nil
	}

	if r.CaseSettings == nil {
		return r, errors.New("specify caseSettings and caseLocation")
	}

	if r.CaseSettings.CaseLocation == "" {
		return r, errors.New("must specify caseLocation for caseSettings")
	}

	var case_description string
	var case_investigator string
	if r.CaseSettings.Case != nil {
		case_description = r.CaseSettings.Case.Description
		case_investigator = r.CaseSettings.Case.Investigator
	}

	r.CaseSettings.Case = &avian.Case{
		Name: r.Name + "-single",
		Directory: fmt.Sprintf("%s/%s-single",
			r.CaseSettings.CaseLocation,
			r.Name,
		),
		Description:  case_description,
		Investigator: case_investigator,
	}

	if r.CaseSettings.CompoundCase == nil || r.CaseSettings.CompoundCase.Directory == "" {
		var compound_description string
		var compound_investigator string
		if r.CaseSettings.CompoundCase != nil {
			compound_description = r.CaseSettings.ReviewCompound.Description
			compound_investigator = r.CaseSettings.ReviewCompound.Investigator
		}

		r.CaseSettings.CompoundCase = &avian.Case{
			Name: r.Name + "-compound",
			Directory: fmt.Sprintf("%s/%s-compound",
				r.CaseSettings.CaseLocation,
				r.Name,
			),
			Description:  compound_description,
			Investigator: compound_investigator,
		}
	}

	if r.CaseSettings.ReviewCompound == nil || r.CaseSettings.ReviewCompound.Directory == "" {
		var review_description string
		var review_investigator string
		if r.CaseSettings.ReviewCompound != nil {
			review_description = r.CaseSettings.ReviewCompound.Description
			review_investigator = r.CaseSettings.ReviewCompound.Investigator
		}

		r.CaseSettings.ReviewCompound = &avian.Case{
			Name: r.Name + "-review",
			Directory: fmt.Sprintf("%s/%s-review",
				r.CaseSettings.CaseLocation,
				r.Name,
			),
			Description:  review_description,
			Investigator: review_investigator,
		}
	}

	return r, nil
}

func readYAML(path string, cfg *Config) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	decoder := yaml.NewDecoder(file)
	return decoder.Decode(cfg)
}

// Get returns data from yml file specified as path
func Get(path string) (*Config, error) {
	var cfg Config
	err := readYAML(path, &cfg)
	return &cfg, err
}
