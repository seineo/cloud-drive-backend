package service

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
	"html/template"
	"path"
	"path/filepath"
	"runtime"
)

type CodeEmailContent struct {
	Code        string
	ProjectName string
	ProjectURL  string
}

type SharedEmailContent struct {
	OwnerName   string
	OwnerEmail  string
	UserBody    string
	Buttons     []EmailButton
	ProjectName string
	ProjectURL  string
}

type EmailButton struct {
	Name string
	Link string
}

func sendEmail(emails []string, subject string, body string) error {
	m := gomail.NewMessage()
	for _, email := range emails {
		m.SetHeader("From", emailConfig.FromEmailUser)
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
	subject := fmt.Sprintf("%s-邮件认证", configs.Project.Name)
	code := GenerateCode(6) // 6-digit verification code
	log.WithFields(logrus.Fields{
		"code": code,
	}).Debug("generated authentication code")

	// load template
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("No caller information")
	}
	tmpl, err := loadTemplate(path.Dir(filename), "code_email.html")
	if err != nil {
		return "", err
	}
	// load data and execute template
	emailContent := CodeEmailContent{
		Code:        code,
		ProjectName: configs.Project.Name,
		ProjectURL:  configs.Project.URL,
	}
	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, emailContent)
	if err != nil {
		return "", err
	}
	//log.WithFields(logrus.Fields{
	//	"body": buf.String(),
	//}).Debug("get code email body")
	// send email using template output
	err = sendEmail([]string{email}, subject, buf.String())
	if err != nil {
		return "", err
	}
	return code, nil
}

// SendShareEmails send share email to given target, load from local template and write with user-given body
func SendShareEmails(ownerName string, ownerEmail string, email string, userBody string, fileNames []string, shareLinks []string) error {
	subject := fmt.Sprintf("%s-文件共享", configs.Project.Name)
	// load template
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("No caller information")
	}
	tmpl, err := loadTemplate(path.Dir(filename), "share_email.html")
	if err != nil {
		return err
	}
	// load data and execute template
	var buttons []EmailButton
	for i := range fileNames {
		buttons = append(buttons, EmailButton{
			Name: fileNames[i],
			Link: shareLinks[i],
		})
	}
	emailContent := SharedEmailContent{
		OwnerName:   ownerName,
		OwnerEmail:  ownerEmail,
		UserBody:    userBody,
		Buttons:     buttons,
		ProjectName: configs.Project.Name,
		ProjectURL:  configs.Project.URL,
	}
	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, emailContent)
	if err != nil {
		return err
	}
	err = sendEmail([]string{email}, subject, buf.String())
	return err
}

func loadTemplate(dir string, filename string) (*template.Template, error) {
	tmplPath := filepath.Join(dir, filename)
	log.WithFields(logrus.Fields{
		"absPath": tmplPath,
	}).Debug("get template absolute path")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return nil, err
	}
	return tmpl, nil
}
