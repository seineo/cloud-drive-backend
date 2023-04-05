package config

import (
	"github.com/sirupsen/logrus"
	"os"
)

func initLogger() *logrus.Logger {
	log := logrus.New()
	log.SetReportCaller(true)
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05", // golang std format has to use this datetime
	})
	log.SetLevel(logrus.DebugLevel)
	log.SetOutput(os.Stdout)
	return log
}
