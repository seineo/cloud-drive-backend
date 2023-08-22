package config

import (
	"github.com/sirupsen/logrus"
	"os"
)

var logger *logrus.Logger

func GetLogger() *logrus.Logger {
	if logger == nil {
		return initLogger()
	}
	return logger
}

func initLogger() *logrus.Logger {
	logger = logrus.New()
	logger.SetReportCaller(true)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05", // golang std format has to use this datetime
	})
	logger.SetLevel(logrus.DebugLevel)
	logger.SetOutput(os.Stdout)
	return logger
}
