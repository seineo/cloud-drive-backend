package middleware

import (
	"CloudDrive/service"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func init() {
	log = service.GetLogger()
}
