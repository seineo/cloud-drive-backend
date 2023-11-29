package service

import "CloudDrive/domain/account/entity"

type AccountService interface {
	NewAccount(email string, nickname string, password string) (*entity.Account, error)
	ChangeEmail(accountID int, newEmail string) error
	ChangeNickname(accountID int, newName string) error
	ChangePassword(accountID int, newPassword string) error
	DeleteAccount(accountID int) error
}
