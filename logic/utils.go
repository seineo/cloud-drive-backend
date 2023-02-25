package logic

import (
	"CloudDrive/config"
	"gopkg.in/gomail.v2"
	"math/rand"
	"time"
)

// Send email with text/html body format
func sendEmail(email string, subject string, HTMLBody string, attachLoc string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", config.FromEmail)
	m.SetHeader("To", email)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", HTMLBody)
	if attachLoc != "" {
		m.Attach(attachLoc)
	}
	d := gomail.NewDialer(config.SMTPHost, config.SMTPPort, config.FromEmailUser, config.FromEmailPsw)
	err := d.DialAndSend(m)
	return err
}

// Generate n-digit verification code for email confirmation
func getVerificationCode(n int) string {
	const digitBytes = "0123456789"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = digitBytes[rand.Intn(len(digitBytes))]
	}
	return string(b)
}
