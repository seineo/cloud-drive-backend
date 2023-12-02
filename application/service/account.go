package service

import (
	"CloudDrive/adapters/http"
	"CloudDrive/domain/account/entity"
	"CloudDrive/domain/account/service"
)

type ApplicationAccount interface {
	Create(user http.UserSignUpRequest) (*entity.Account, error)
}

type applicationAccount struct {
	accountService service.AccountService
}

func (a *applicationAccount) Create(user http.UserSignUpRequest) (*entity.Account, error) {
	account, err := a.accountService.NewAccount(user.Email, user.Name, user.Password)
	if err != nil {
		return nil, err
	}
	return account, nil
}

func NewApplicationAccount(acc service.AccountService) ApplicationAccount {
	return &applicationAccount{accountService: acc}
}
