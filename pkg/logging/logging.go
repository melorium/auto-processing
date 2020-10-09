package logging

import (
	"fmt"
	"os"
	"time"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Service interface {
	Clean(from time.Time)
	Get(name string) (*zap.Logger, error)
}

type svc struct {
	LogPath string
	Holder  map[string]*Logholder
}

type Logholder struct {
	lastUsed time.Time
	lumber   *lumberjack.Logger
	logger   *zap.Logger
	file     *os.File
}

func New(logPath string) Service {
	return svc{
		LogPath: logPath,
		Holder:  make(map[string]*Logholder),
	}
}

// Clean the log-holder of inactive logs
// from a specified time
func (s svc) Clean(from time.Time) {
	for name, holder := range s.Holder {
		if holder.lastUsed.After(from) {
			continue
		}

		holder.logger = nil
		holder.lumber.Close()
		holder.lumber = nil
		holder.file.Close()
		holder.file = nil
		delete(s.Holder, name)
	}
}

// Get the specified log by name
// - if not in memory, it will be created/opened
func (s svc) Get(name string) (*zap.Logger, error) {
	if log, ok := s.Holder[name]; ok {
		log.lastUsed = time.Now()
		return log.logger, nil
	}

	return s.open(name)
}

// open will create or open the log specified by name
func (s svc) open(logName string) (*zap.Logger, error) {
	log, err := os.OpenFile(s.LogPath+logName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("error opening log-file %s: %v", logName, err)
	}

	lumberjackLogger := &lumberjack.Logger{
		Filename:   log.Name(),
		MaxSize:    0, // megabytes
		MaxBackups: 3,
		MaxAge:     1, //days
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(lumberjackLogger),
		zap.DebugLevel,
	)

	logger := zap.New(core)
	s.Holder[logName] = &Logholder{
		lastUsed: time.Now(),
		logger:   logger,
		lumber:   lumberjackLogger,
		file:     log,
	}
	return logger, nil
}
