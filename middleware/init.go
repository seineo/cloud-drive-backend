package middleware

import (
	"CloudDrive/config"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func init() {
	log = config.GetConfig().Log
}
