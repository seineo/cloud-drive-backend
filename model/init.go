package model

import (
	"CloudDrive/config"
	"fmt"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB
var log *logrus.Logger
var configs config.Config

func init() {
	log = config.GetLogger()
	// mysql
	configs = config.LoadConfig("./config")
	mysqlConfig := configs.Database
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
	// SetUpJoinTable should set before AutoMigrate
	err = db.SetupJoinTable(&Directory{}, "Files", &DirectoryFile{})
	if err != nil {
		log.WithError(err).Fatal("fail to set up join table")
	}
	// auto migration for models, for example creating tables automatically
	err = db.AutoMigrate(&User{}, &File{}, &Directory{})
	if err != nil {
		log.WithError(err).Fatal("fail to auto migrate models")
	}

}
