package service

import (
	"account/adapters/http/types"
	"account/domain/account/entity"
	"account/domain/account/service"
	"common/slugerror"
	"github.com/alexedwards/argon2id"
)

type ApplicationAccount interface {
	Create(user types.AccountSignUpRequest) (*entity.Account, error)
	Login(user types.AccountLoginRequest) (*entity.Account, error)
	Get(accountID uint) (*entity.Account, error)
	Update(accountID uint, user types.AccountUpdateRequest) error
	Delete(accountID uint) error
}

type applicationAccount struct {
	accountService service.AccountService
}

func (a *applicationAccount) Get(accountID uint) (*entity.Account, error) {
	return a.accountService.GetAccountByID(accountID)
}

func (a *applicationAccount) Login(user types.AccountLoginRequest) (*entity.Account, error) {
	// 查看邮箱对应账号是否存在
	account, err := a.accountService.GetAccountByEmail(user.Email)
	if err != nil {
		return nil, err
	}
	// 查看账号密码是否匹配
	match, err := argon2id.ComparePasswordAndHash(user.Password, account.GetPassword())
	if err != nil {
		return nil, err
	}
	if !match {
		return nil, slugerror.NewSlugError(slugerror.ErrUnauthorized, "invalid password", "password not match")
	}
	return account, nil
}

func (a *applicationAccount) Create(user types.AccountSignUpRequest) (*entity.Account, error) {
	account, err := a.accountService.NewAccount(user.Email, user.Nickname, user.Password)
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (a *applicationAccount) Update(accountID uint, user types.AccountUpdateRequest) error {
	if len(user.Email) > 0 {
		if err := a.accountService.ChangeEmail(accountID, user.Email); err != nil {
			return err
		}
	} else if len(user.Nickname) > 0 {
		if err := a.accountService.ChangeNickname(accountID, user.Nickname); err != nil {
			return err
		}
	} else if len(user.Password) > 0 {
		if err := a.accountService.ChangePassword(accountID, user.Password); err != nil {
			return err
		}
	} else {
		return slugerror.NewSlugError(slugerror.ErrInvalidInput,
			"empty request data", "The request needs to have a field that is not empty")
	}
	return nil
}

func (a *applicationAccount) Delete(accountID uint) error {
	return a.accountService.DeleteAccount(accountID)
}

func NewApplicationAccount(acc service.AccountService) ApplicationAccount {
	return &applicationAccount{accountService: acc}
}
