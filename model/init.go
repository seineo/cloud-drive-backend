package model

import (
	"CloudDrive/config"
	"CloudDrive/service"
	"fmt"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB
var log *logrus.Logger

func init() {
	log = service.GetLogger()
	// mysql
	mysqlConfig := config.GetConfig().Storage.MySQL
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
