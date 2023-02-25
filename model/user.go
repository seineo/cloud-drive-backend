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
	Name       string `form:"name"`
	Email      string `form:"email"`
	Password   string `form:"password"`
}

var db *gorm.DB
var log *logrus.Logger

func init() {
	log = service.GetLogger()
	// mysql
	mysqlConfig := config.GetConfig().MySQL
	dsn := fmt.Sprintf("%s:%s@%s(%s)/%s?parseTime=true",
		mysqlConfig.User, mysqlConfig.Password, mysqlConfig.Protocol, mysqlConfig.Address, mysqlConfig.Database)
	log.Debugf("sdn for mysql connnection : %s", dsn)
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

func CreateUser(user *User) (uint, error) {
	err := db.Create(user).Error
	return user.ID, err
}

func GetUserByID(id uint) (*User, error) {
	user := &User{}
	err := db.First(user, id).Error
	return user, err
}

func DeleteUser(email string) error {
	return db.Where("email = ?", email).Delete(&User{}).Error
}
