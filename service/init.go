package service

import (
	"CloudDrive/config"
	"github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
)

var log *logrus.Logger
var emailConfig *config.EmailConfig
var emailSender gomail.SendCloser

func init() {
	log = config.GetConfig().Log
	emailConfig = config.GetConfig().Email
	d := gomail.NewDialer(emailConfig.SMTPHost, emailConfig.SMTPPort,
		emailConfig.FromEmailUser, emailConfig.FromEmailPsw)
	var err error
	emailSender, err = d.Dial()
	if err != nil {
		log.WithError(err).Fatal("failed to dial smtp server")
	}
}
