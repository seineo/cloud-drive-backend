package model

import (
	"CloudDrive/config"
	"CloudDrive/service"
	"fmt"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model        // embeds id, create at and update at timestamps
	Name       string `form:"name" binding:"required"`
	Email      string `form:"email" binding:"required,email"`
	Password   string `form:"password" binding:"required"`
}

var db *gorm.DB
var log *logrus.Logger

func init() {
	log = service.GetLogger()
	// mysql
	mysqlConfig := config.GetConfig().MySQL
	dsn := fmt.Sprintf("%s:%s@%s(%s)/%s?parseTime=true",
		mysqlConfig.User, mysqlConfig.Password, mysqlConfig.Protocol, mysqlConfig.Address, mysqlConfig.Database)
	log.WithFields(logrus.Fields{
		"dsn": dsn,
	}).Debug("check dsn")

	dbConn, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.WithError(err).Fatal("fail to connect mysql database")
	}
	db = dbConn
	err = db.AutoMigrate(&User{})
	if err != nil {
		log.WithError(err).Error("fail to auto migrate model user")
	}
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
