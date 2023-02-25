package service

import (
	"github.com/sirupsen/logrus"
	"os"
)

var log *logrus.Logger

func GetLogger() *logrus.Logger {
	if log == nil {
		InitLogger()
	}
	return log
}

func InitLogger() {
	log = logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetLevel(logrus.DebugLevel)
	log.SetOutput(os.Stdout)
}
