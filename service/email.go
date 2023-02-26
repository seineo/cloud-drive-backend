package service

import (
	"CloudDrive/config"
	"bytes"
	"github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
	"html/template"
)

var emailConfig *config.EmailConfig
var emailSender gomail.SendCloser

type CodeEmailContent struct {
	Code        string
	ProjectName string
	ProjectURL  string
}

func init() {
	log = GetLogger()
	emailConfig = config.GetConfig().Email
	d := gomail.NewDialer(emailConfig.SMTPHost, emailConfig.SMTPPort,
		emailConfig.FromEmailUser, emailConfig.FromEmailPsw)
	var err error
	emailSender, err = d.Dial()
	if err != nil {
		log.WithError(err).Fatal("failed to dial smtp server")
	}
}

func sendEmail(emails []string, subject string, body string) error {
	m := gomail.NewMessage()
	for _, email := range emails {
		m.SetHeader("From", emailConfig.FromEmail)
		m.SetHeader("To", email)
		m.SetHeader("Subject", subject)
		m.SetBody("text/html", body)

		if err := gomail.Send(emailSender, m); err != nil {
			return err
		}
		log.WithFields(logrus.Fields{
			"to": email,
		}).Info("successfully sent email")
		m.Reset()
	}
	return nil
}

// SendCodeEmail sends verification code to registering users
func SendCodeEmail(email string) (string, error) {
	subject := "Email Confirmation"
	code := GenerateCode(6) // 6-digit verification code
	log.WithFields(logrus.Fields{
		"code": code,
	}).Debug("generated authentication code")

	// get template
	tmpl, err := template.ParseFiles("./code_email.html")
	if err != nil {
		return "", err
	}

	// load data and execute template
	emailContent := CodeEmailContent{
		Code:        code,
		ProjectName: config.GetConfig().ProjectName,
		ProjectURL:  config.GetConfig().ProjectURL,
	}
	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, emailContent)
	if err != nil {
		return "", err
	}
	log.WithFields(logrus.Fields{
		"body": buf.String(),
	}).Debug("get code email body")
	// send email using template output
	err = sendEmail([]string{email}, subject, buf.String())
	if err != nil {
		return "", err
	}
	return code, nil
}
