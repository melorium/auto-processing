package config

import (
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

func (cfg *Config) Validate() error {
	log.Println("Validating config")
	if _, err := os.Stat(cfg.Server.NuixPath); os.IsNotExist(err) {
		log.Println("Missing nuix-path")
		return err
	} 
	
	if _, err := os.Stat(cfg.Nuix.ProcessProfilePath); os.IsNotExist(err) {
		log.Println("cant find path for process-profile")
		return err
	} 

	/*
	if _, err := os.Stat(cfg.Nuix.Settings); os.IsNotExist(err) {
		log.Println("cant find path for the settings")	
		return err
	}   
	*/

	
	return nil
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