package entity

import (
	"CloudDrive/common/validation"
	"errors"
	"fmt"
	"github.com/alexedwards/argon2id"
)

type Account struct {
	id       uint
	email    string
	nickname string
	password string
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
	// 使用 argon 加密密码
	hashedPassword, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return nil, err
	}
	return &Account{
		email:    email,
		nickname: nickname,
		password: hashedPassword,
	}, nil
}

// NewAccountWithID 仅仅用于测试，不要用于账户初始化，因为本函数不做参数验证
func NewAccountWithID(id uint, email string, nickname string, password string) *Account {
	hashedPassword, _ := argon2id.CreateHash(password, argon2id.DefaultParams)
	return &Account{
		id:       id,
		email:    email,
		nickname: nickname,
		password: hashedPassword,
	}
}

// UnmarshallAccount 从仓储实体映射回来领域实体，因为本函数不做参数验证和参数转换
func UnmarshallAccount(id uint, email string, nickname string, password string) *Account {
	return &Account{
		id:       id,
		email:    email,
		nickname: nickname,
		password: password,
	}
}

func (a *Account) GetID() uint {
	return a.id
}

func (a *Account) GetEmail() string {
	return a.email
}

func (a *Account) GetNickname() string {
	return a.nickname
}

func (a *Account) GetPassword() string {
	return a.password
}

func (a *Account) UpdateEmail(newEmail string) error {
	err := validation.CheckEmail(newEmail)
	if err != nil {
		return err
	}
	a.email = newEmail
	return nil
}

func (a *Account) UpdateNickname(fc FactoryConfig, newName string) error {
	err := validation.CheckRegexMatch(fc.NicknameRegex, newName)
	if err != nil {
		return err
	}
	a.nickname = newName
	return nil
}

func (a *Account) UpdatePassword(fc FactoryConfig, newPassword string) error {
	err := validation.CheckRegexMatch(fc.PasswordRegex, newPassword)
	if err != nil {
		return err
	}
	a.password = newPassword
	return nil
}
