package model

import (
	"CloudDrive/request"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name        string
	Email       string
	Password    string
	RootHash    string
	Directories []Directory
}

// CreateUser stores user info in table `users`, and inserts root directory metadata into table `directories`
func CreateUser(userRequest *request.UserSignUpRequest) (*User, error) {
	user := &User{
		Name:     userRequest.Name,
		Email:    userRequest.Email,
		Password: userRequest.Password,
		RootHash: userRequest.RootHash,
	}
	user.Directories = append(user.Directories, Directory{
		Hash:   user.RootHash,
		UserID: user.ID,
		Name:   "我的云盘",
	})
	if err := db.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
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
