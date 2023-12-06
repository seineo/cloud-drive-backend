package service

import (
	config2 "CloudDrive/common/config"
	"github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
)

var log *logrus.Logger
var configs config2.Config
var emailConfig config2.EmailConfig
var emailSender gomail.SendCloser

func init() {
	log = config2.GetLogger()
	log.Info("service calls config")
	configs = config2.LoadConfig([]string{"config", "../config"})
	emailConfig = configs.Email
	d := gomail.NewDialer(emailConfig.SMTPHost, emailConfig.SMTPPort,
		emailConfig.User, emailConfig.Password)
	var err error
	emailSender, err = d.Dial()
	if err != nil {
		log.WithError(err).Fatal("failed to dial smtp server")
	}
}
