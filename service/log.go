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
	log.SetReportCaller(true)
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05", // golang std format has to use this datetime
	})
	log.SetLevel(logrus.DebugLevel)
	log.SetOutput(os.Stdout)
}
