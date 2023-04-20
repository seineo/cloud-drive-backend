package service

import (
	"CloudDrive/model"
	"time"
)

func GetShareExpiredTimePtr(expiredTime string) *time.Time {
	// TODO parse expired time in string format
	curTime := time.Now()
	return &curTime
}

func GetSharedUserIDPtr(user *model.User) *uint {
	if user != nil {
		return &user.ID
	} else {
		return nil
	}
}

func GetPasswordPtr(password string) *string {
	if password == "" {
		return nil
	} else {
		return &password
	}
}
