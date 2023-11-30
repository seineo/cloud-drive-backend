package service

import (
	"CloudDrive/common/validation"
	"CloudDrive/domain/account/entity"
	"CloudDrive/domain/account/repository"
	"errors"
)

type AccountService interface {
	NewAccount(email string, nickname string, password string) (*entity.Account, error)
	ChangeEmail(accountID int, newEmail string) error
	ChangeNickname(accountID int, newName string) error
	ChangePassword(accountID int, newPassword string) error
	DeleteAccount(accountID int) error
}

type accountService struct {
	accountRepo repository.AccountRepo
	accountFc   entity.FactoryConfig
}

var EmailUsedError = errors.New("email has already been used")

func (svc *accountService) checkEmailNotUsed(email string) error {
	account, err := svc.accountRepo.GetByEmail(email)
	if err != nil {
		return err
	}
	if account != nil {
		return EmailUsedError
	}
	return nil
}

func (svc *accountService) NewAccount(email string, nickname string, password string) (*entity.Account, error) {
	factory, err := entity.NewFactory(svc.accountFc)
	if err != nil {
		return nil, err
	}
	account, err := factory.NewAccount(email, nickname, password)
	if err != nil {
		return nil, err
	}
	err = svc.checkEmailNotUsed(email)
	if err != nil {
		return nil, err
	}
	newAccount, err := svc.accountRepo.Create(*account)
	if err != nil {
		return nil, err
	}
	return newAccount, nil
}

func (svc *accountService) ChangeEmail(accountID int, newEmail string) error {
	err := validation.CheckEmail(newEmail)
	if err != nil {
		return err
	}
	err = svc.checkEmailNotUsed(newEmail)
	if err != nil {
		return err
	}
	account, err := svc.accountRepo.Get(accountID)
	if err != nil {
		return err
	}
	account.Email = newEmail
	_, err = svc.accountRepo.Update(*account)
	if err != nil {
		return err
	}
	return nil
}

func (svc *accountService) ChangeNickname(accountID int, newName string) error {
	err := validation.CheckRegexMatch(svc.accountFc.NicknameRegex, newName)
	if err != nil {
		return err
	}
	account, err := svc.accountRepo.Get(accountID)
	if err != nil {
		return err
	}
	account.Nickname = newName
	_, err = svc.accountRepo.Update(*account)
	if err != nil {
		return err
	}
	return nil
}

func (svc *accountService) ChangePassword(accountID int, newPassword string) error {
	err := validation.CheckRegexMatch(svc.accountFc.PasswordRegex, newPassword)
	if err != nil {
		return err
	}
	account, err := svc.accountRepo.Get(accountID)
	if err != nil {
		return err
	}
	account.Password = newPassword
	_, err = svc.accountRepo.Update(*account)
	if err != nil {
		return err
	}
	return nil
}

func (svc *accountService) DeleteAccount(accountID int) error {
	return svc.accountRepo.Delete(accountID)
}

func NewAccountService(accountRepo repository.AccountRepo, fc entity.FactoryConfig) AccountService {
	return &accountService{
		accountRepo: accountRepo,
		accountFc:   fc,
	}
}
