package logger

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/maleck13/scm-go/config"
	"io"
)

var Logger = logrus.New()
var logFilePath = "/var/log/feedhenry/fh-scm/"

func directoryExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	return false, err
}

func InitLogger(loggers config.Logger) *logrus.Logger {
	logger := loggers.Streams[0]
	Logger.Formatter = new(logrus.JSONFormatter)

	// Check if we can use the standard log directory. If not
	// fall back to tmp
	if result, _ := directoryExists(logFilePath); !result {
		logFilePath = "/tmp/"
	}

	// Use the 'Stream' value from the fh-scm config as the filename
	// for now
	logFileName := logFilePath + logger.Stream

	f, err := os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		Logger.Fatal(err)
	}

	// Log to file and stdout
	Logger.Out = io.MultiWriter(f, os.Stdout)

	switch logger.Level {
	case "info":
		Logger.Level = logrus.InfoLevel
		break
	case "error":
		Logger.Level = logrus.ErrorLevel
		break
	case "debug":
		Logger.Level = logrus.DebugLevel
		break

	}
	return Logger
}
