package logger

import (
	"os"

	"github.com/fheng/scm-go/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/fheng/scm-go/config"
)

//TODO write to file and stdout

var Logger = logrus.New()

func InitLogger(loggers []config.Logger) *logrus.Logger {
	logger := loggers[0]
	Logger.Formatter = new(logrus.JSONFormatter)
	f, err := os.OpenFile(logger.Filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		Logger.Fatal(err)
	}
	Logger.Out = f
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
