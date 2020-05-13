package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/shiena/ansicolor"
	"github.com/sirupsen/logrus"

	"gopkg.in/yaml.v2"
)

type Script struct {
	Logger *logrus.Logger
}

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
	Xmx string `json:"xmx"`
	Workers string `yaml:"workers"`
	ProcessProfilePath string `yaml:"process_profile_path"`
	ProcessProfileName string `yaml:"process_profile_name"`
	Settings string `yaml:"settings"`
	Switches []string `yaml:"switches"`
}

func (s *Script) readYAML(path string, cfg *Config) {
	log := s.Logger
	file, err := os.Open(path)
	if err != nil {
		log.Errorln("Missing config-file -", err)
		os.Exit(2)
	}

	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(cfg)
	if err != nil {
		log.Errorln("Failed to decode config-file -", err)
		os.Exit(2)
	}
}

// getConfig returns data from config.yml
func (s *Script) getConfig(path string) *Config {
	var cfg Config
	s.readYAML(path, &cfg)
	return &cfg
}

type myFormatter struct {
    logrus.TextFormatter
}

func (f *myFormatter) Format(entry *logrus.Entry) ([]byte, error) {
// this whole mess of dealing with ansi color codes is required if you want the colored output otherwise you will lose colors in the log levels
    var levelColor int
    switch entry.Level {
    case logrus.DebugLevel, logrus.TraceLevel:
        levelColor = 31 // gray
    case logrus.WarnLevel:
        levelColor = 33 // yellow
    case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
        levelColor = 31 // red
    default:
        levelColor = 36 // blue
    }
    return []byte(fmt.Sprintf("[%s] - \x1b[%dm%s\x1b[0m - %s\n", entry.Time.Format(f.TimestampFormat), levelColor, strings.ToUpper(entry.Level.String()), entry.Message)), nil
}

func main() {
	f, _ := os.OpenFile("log.txt", os.O_CREATE|os.O_WRONLY, 0777)
    log := &logrus.Logger{
        Out:   io.MultiWriter(os.Stderr, f),
        Level: logrus.InfoLevel,
        Formatter: &myFormatter{logrus.TextFormatter{
            FullTimestamp:          true,
            TimestampFormat:        "2006-01-02 15:04:05",
            ForceColors:            true,
			DisableLevelTruncation: false,
			
        },
        },
	}
	log.Out = ansicolor.NewAnsiColorWriter(os.Stdout)
	log.Info("Initializing script")
	script := Script{Logger: log}

	cfg := script.getConfig(".\\config.yml")
	program := cfg.Server.NuixPath + "\\nuix_console.exe"

	log.Info("Validating config")
	if _, err := os.Stat(cfg.Server.NuixPath); os.IsNotExist(err) {
		log.WithFields(logrus.Fields{
			"where": "config",
			"exception": err,
		}).Error("Cant find nuix path")	
		return
	} 
	
	if _, err := os.Stat(cfg.Nuix.ProcessProfilePath); os.IsNotExist(err) {
		log.WithFields(logrus.Fields{
			"where": "config",
			"exception": err,
		}).Error("cant find path for process-profile")	
		return
	} 

	if _, err := os.Stat(cfg.Nuix.Settings); os.IsNotExist(err) {
		log.WithFields(logrus.Fields{
			"where": "config",
			"exception": err,
		}).Error("cant find path for the settings")	
		return
	} 

	path, err := os.Getwd()
	if err != nil {
		log.WithFields(logrus.Fields{
			"where": "os",
			"exception": err,
		}).Error("cant get working directory")	
		return
	}

	cmd := exec.Command(
		program,
		"-Xmx" + cfg.Nuix.Xmx,
		"-Dnuix.registry.servers=" + cfg.Server.NmsAddress,
		"-licencesourcetype",
		"server",
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
		cfg.Nuix.Settings,
		strings.Join(cfg.Nuix.Switches , " "),
	)
	
	cmd.Dir = cfg.Server.NuixPath
	cmd.Env = append(os.Environ(),
		"NUIX_USERNAME=" + cfg.Server.Username,
		"NUIX_PASSWORD=" + cfg.Server.Password,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	log.Info("Executing Nuix with command: ", cmd)

	if err := cmd.Start(); err != nil {
		log.WithFields(logrus.Fields{
			"where": "nuix",
			"exception": err,
		}).Error("Failed to run program")	
		return
	}
	if err := cmd.Wait(); err != nil {
		log.WithFields(logrus.Fields{
			"where": "nuix",
			"exception": err,
		}).Error("Failed to run program")	
		return
	}
}