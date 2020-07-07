package log

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

func Get(source, cfg string) (*logrus.Entry, *os.File) {
	dt := time.Now()
	timestamp := dt.Format("20060102")
	filename := fmt.Sprintf("%s%s.log", source, timestamp)

	// create log-dir
	if _, err := os.Stat("./logs"); os.IsNotExist(err) {
		os.Mkdir("./logs", os.ModePerm)
	}

	// create log-file with read-write & create-permissions
	logFile, err := os.OpenFile("./logs/"+filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Failed to open logfile: %v", err)
		os.Exit(2)
	}

	formatter := &logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "msg",
			logrus.FieldKeyTime:  "time",
		},
	}

	logger := &logrus.Logger{
		Out:       io.MultiWriter(os.Stdout, logFile),
		Level:     logrus.DebugLevel,
		Formatter: formatter,
	}

	log := logger.WithFields(logrus.Fields{"src": source, "config": cfg})

	return log, logFile
}
