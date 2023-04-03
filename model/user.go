package model

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model        // embeds id, create at and update at timestamps
	Name       string `form:"name" binding:"required"`
	Email      string `form:"email" binding:"required,email"`
	Password   string `form:"password" binding:"required"`
}

func CreateUser(user *User) error {
	return db.Create(user).Error
}

func GetUserByID(id uint) (*User, error) {
	user := &User{}
	err := db.First(user, id).Error
	return user, err
}

func GetUserByEmail(email string) (*User, error) {
	user := &User{}
	err := db.Where("email = ?", email).First(user).Error
	return user, err
}

func DeleteUser(email string) error {
	return db.Where("email = ?", email).Delete(&User{}).Error
}
