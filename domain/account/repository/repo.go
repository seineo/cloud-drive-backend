package repository

import "CloudDrive/domain/account/entity"

type AccountRepo interface {
	Get(accountID int) (*entity.Account, error)
	GetByEmail(email string) (*entity.Account, error)
	Create(account entity.Account) (*entity.Account, error)
	Update(account entity.Account) (*entity.Account, error)
	Delete(accountID int) error
}
