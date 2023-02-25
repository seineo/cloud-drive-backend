package service

import (
	"bytes"
	"errors"
	"net/mail"
	"text/template"
)

var UserNameError = errors.New("user name should not be empty")
var EmailFormatError = errors.New("email format is not valid")

// CheckUser checks the validity of name and email
func CheckUser(name string, email string) error {
	if name == "" {
		return UserNameError
	}
	if !isEmailValid(email) {
		return EmailFormatError
	}
	return nil
}

func isEmailValid(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// SendCodeEmail sends verification code to registering users
func SendCodeEmail(email string) (string, error) {
	subject := "Email Confirmation"
	code := getVerificationCode(6) // 6-digit verification code
	body, err := getBody(code)
	if err != nil {
		return "", err
	}
	err = sendEmail(email, subject, body, "")
	if err != nil {
		return "", err
	}
	return code, nil
}

func getBody(code string) (string, error) {
	tmpl, err := template.ParseFiles("../front-end/templates/code-email.html")
	if err != nil {
		return "", err
	}
	var output bytes.Buffer
	if err := tmpl.Execute(&output, code); err != nil {
		return "", err
	}
	return output.String(), nil
}
