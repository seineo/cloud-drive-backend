package entity

import (
	"errors"
	"fmt"
	"net/mail"
	"regexp"
)

type Account struct {
	ID       int
	Email    string
	Nickname string
	Password string
}

type FactoryConfig struct {
	NicknameRegex string
	PasswordRegex string
}

type Factory struct {
	fc FactoryConfig
}

func (fc *FactoryConfig) validateConfig() error {
	var err error
	if len(fc.NicknameRegex) == 0 {
		err = errors.Join(err, fmt.Errorf("昵称的正则表达式不应为空"))
	}
	if len(fc.PasswordRegex) == 0 {
		err = errors.Join(err, fmt.Errorf("密码的正则表达式不应为空"))
	}
	return err
}

func NewFactory(fc FactoryConfig) (*Factory, error) {
	if err := fc.validateConfig(); err != nil {
		return nil, err
	}
	return &Factory{fc}, nil
}

func (f *Factory) validate(email string, nickname string, password string) error {
	var err error
	_, mailErr := mail.ParseAddress(email)
	if mailErr != nil {
		err = errors.Join(err, fmt.Errorf("email is no valid: %w", mailErr))
	}
	nicknamePattern := regexp.MustCompile(f.fc.NicknameRegex)
	if !nicknamePattern.MatchString(nickname) {
		err = errors.Join(err, fmt.Errorf("nickname is no valid"))
	}
	passwordPattern := regexp.MustCompile(f.fc.PasswordRegex)
	if !passwordPattern.MatchString(password) {
		err = errors.Join(err, fmt.Errorf("password is no valid"))
	}
	return err
}

func (f *Factory) NewAccount(email string, nickname string, password string) (*Account, error) {
	if err := f.validate(email, nickname, password); err != nil {
		return nil, err
	}
	return &Account{
		Email:    email,
		Nickname: nickname,
		Password: password,
	}, nil
}
