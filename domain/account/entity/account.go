package entity

import (
	"CloudDrive/common/validation"
	"errors"
	"fmt"
)

type Account struct {
	ID       uint
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
		err = errors.Join(err, fmt.Errorf("regex for nickname should not be empty"))
	}
	if len(fc.PasswordRegex) == 0 {
		err = errors.Join(err, fmt.Errorf("regex for password should not be empty"))
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
	mailErr := validation.CheckEmail(email)
	if mailErr != nil {
		err = errors.Join(err, mailErr)
	}
	nicknameErr := validation.CheckRegexMatch(f.fc.NicknameRegex, nickname)
	if nicknameErr != nil {
		err = errors.Join(err, nicknameErr)
	}
	passwordErr := validation.CheckRegexMatch(f.fc.PasswordRegex, password)
	if passwordErr != nil {
		err = errors.Join(err, passwordErr)
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
