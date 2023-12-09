package service

import (
	"account/domain/account/entity"
	"account/domain/account/repository"
	"account/infrastructure/repo"
	"common/slugerror"
	"errors"
)

var EmailUsedError = slugerror.NewSlugError(slugerror.ErrConflict, "resource conflict", "email has already been used")

type AccountService interface {
	NewAccount(email string, nickname string, password string) (*entity.Account, error)
	GetAccountByID(accountID uint) (*entity.Account, error)
	GetAccountByEmail(email string) (*entity.Account, error) // 找不到账号时不报错，Account为空
	ChangeEmail(accountID uint, newEmail string) error
	ChangeNickname(accountID uint, newName string) error
	ChangePassword(accountID uint, newPassword string) error
	DeleteAccount(accountID uint) error
}

type accountService struct {
	accountRepo repository.AccountRepo
	accountFc   entity.AccountFactoryConfig
}

func (svc *accountService) checkEmailNotUsed(email string) error {
	_, err := svc.accountRepo.GetByEmail(email)
	if err != nil {
		if errors.Is(err, repo.RecordNotFoundError) { // 没找到说明没有人使用，则不返回错误
			return nil
		} else {
			return err
		}
	}
	return EmailUsedError
}

func (svc *accountService) NewAccount(email string, nickname string, password string) (*entity.Account, error) {
	factory, err := entity.NewAccountFactory(svc.accountFc)
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

func (svc *accountService) GetAccountByID(accountID uint) (*entity.Account, error) {
	return svc.accountRepo.Get(accountID)
}

func (svc *accountService) GetAccountByEmail(email string) (*entity.Account, error) {
	return svc.accountRepo.GetByEmail(email)
}

func (svc *accountService) ChangeEmail(accountID uint, newEmail string) error {
	err := svc.checkEmailNotUsed(newEmail)
	if err != nil {
		return err
	}
	acc, err := svc.accountRepo.Get(accountID)
	if err != nil {
		return err
	}
	err = acc.UpdateEmail(newEmail)
	if err != nil {
		return err
	}
	_, err = svc.accountRepo.Update(*acc)
	if err != nil {
		return err
	}
	return nil
}

func (svc *accountService) ChangeNickname(accountID uint, newName string) error {
	acc, err := svc.accountRepo.Get(accountID)
	if err != nil {
		return err
	}
	err = acc.UpdateNickname(svc.accountFc, newName)
	if err != nil {
		return err
	}
	_, err = svc.accountRepo.Update(*acc)
	if err != nil {
		return err
	}
	return nil
}

func (svc *accountService) ChangePassword(accountID uint, newPassword string) error {
	acc, err := svc.accountRepo.Get(accountID)
	if err != nil {
		return err
	}
	err = acc.UpdatePassword(svc.accountFc, newPassword)
	if err != nil {
		return err
	}
	_, err = svc.accountRepo.Update(*acc)
	if err != nil {
		return err
	}
	return nil
}

func (svc *accountService) DeleteAccount(accountID uint) error {
	return svc.accountRepo.Delete(accountID)
}

func NewAccountService(accountRepo repository.AccountRepo, fc entity.AccountFactoryConfig) AccountService {
	return &accountService{
		accountRepo: accountRepo,
		accountFc:   fc,
	}
}
