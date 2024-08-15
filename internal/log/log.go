package log

import (
	"path/filepath"
	"sync"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	maxSize    = 500
	maxAge     = 30
	maxBackups = 5
)

var (
	loggers  []*logrus.Logger
	logMutex sync.Mutex
)

func init() {
	loggers = make([]*logrus.Logger, 0)
}

func NewLogger(logPath string) *logrus.Logger {
	logMutex.Lock()
	defer logMutex.Unlock()

	logPath = filepath.Join("logs", logPath)

	for _, log := range loggers {
		if log.Out.(*lumberjack.Logger).Filename == logPath {
			return log
		}
	}

	log := logrus.New()

	output := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
		Compress:   true,
		LocalTime:  true,
	}

	log.SetOutput(output)
	log.SetLevel(logrus.DebugLevel)

	loggers = append(loggers, log)
	return log
}
