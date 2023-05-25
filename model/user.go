package model

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model         // embeds id, create at and update at timestamps
	Name        string `json:"name" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required"`
	RootHash    string `json:"rootHash" binding:"required"`
	Directories []Directory
}

// CreateUser stores user info in table `users`, and inserts root directory metadata into table `directories`
func CreateUser(user *User) error {
	// note tha we should use tx instead of db in the scope of transaction
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		err := tx.Create(&Directory{
			Hash:       user.RootHash,
			UserID:     user.ID,
			Name:       "我的云盘",
			ParentHash: nil,
		}).Error
		if err != nil {
			return err
		}
		return nil
	})
	return err
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
