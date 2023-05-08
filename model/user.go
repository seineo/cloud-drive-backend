package model

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model        // embeds id, create at and update at timestamps
	Name       string `json:"name" binding:"required"`
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required"`
	RootHash   string `json:"rootHash" binding:"required"`
}

func CreateUser(user *User) error {
	return db.Create(user).Error
}

func GetUserByID(id uint) (*User, error) {
	var user User
	if err := db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func GetUserByEmail(email string) (*User, error) {
	var user User
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
