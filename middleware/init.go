package middleware

import (
	"CloudDrive/common/config"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func init() {
	log = config.GetLogger()
}
