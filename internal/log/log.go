package log

import (
	"io"
	"os"
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

func NewLogger(logPath string) *logrus.Logger {
	logMutex.Lock()
	defer logMutex.Unlock()

	logPath = filepath.Join("logs", logPath)

	for _, log := range loggers {
		if log.Out == logFileOutput(logPath) {
			return log
		}
	}

	log := logrus.New()

	fileOutput := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
		Compress:   true,
		LocalTime:  true,
	}

	multiWriter := io.MultiWriter(fileOutput, os.Stdout)

	log.SetOutput(multiWriter)
	log.SetLevel(logrus.DebugLevel)

	loggers = append(loggers, log)
	return log
}

func logFileOutput(path string) io.Writer {
	return &lumberjack.Logger{
		Filename: path,
	}
}
