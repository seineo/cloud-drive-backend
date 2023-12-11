package entity

import (
	"common/validation"
	"fmt"
)

type Email struct {
	sender     string
	recipients []string
	subject    string
	body       string
}

func (e *Email) Sender() string {
	return e.sender
}

func (e *Email) Recipients() []string {
	return e.recipients
}

func (e *Email) Subject() string {
	return e.subject
}

func (e *Email) Body() string {
	return e.body
}

func NewEmail(sender string, recipients []string, subject string, body string) (*Email, error) {
	if err := validation.CheckEmail(sender); err != nil {
		return nil, err
	}
	if len(recipients) == 0 {
		return nil, fmt.Errorf("email should have at least one recipients")
	}
	for _, recipient := range recipients {
		if err := validation.CheckEmail(recipient); err != nil {
			return nil, err
		}
	}
	return &Email{
		sender:     sender,
		recipients: recipients,
		subject:    subject,
		body:       body,
	}, nil
}
