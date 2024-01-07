package infrastructure

import (
	"email/domain/infrastructure"
	"github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
)

type GoMailer struct {
	dialer *gomail.Dialer
}

func (g *GoMailer) SendEmail(from string, to string, subject string, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)
	if err := g.dialer.DialAndSend(m); err != nil {
		return err
	}
	logrus.Infof("email sent to %s", to)
	return nil
}

func NewGoMailer(dialer *gomail.Dialer) infrastructure.EmailSender {
	return &GoMailer{dialer: dialer}
}
