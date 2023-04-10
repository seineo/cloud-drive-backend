package main

import (
	"CloudDrive/config"
	"CloudDrive/handler"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"os"
	"runtime"
)

var log *logrus.Logger

func init() {
	_config := config.GetConfig()
	log = _config.Log
	// create file storage directories if they don't exist
	err := os.MkdirAll(_config.Storage.DiskStoragePath, 0750)
	if err != nil {
		log.Fatal("failed to create file storage directory")
	}
	err = os.MkdirAll(_config.Storage.DiskTempStoragePath, 0750)
	if err != nil {
		log.Fatal("failed to create temporal file storage directory")
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU() * 2)

	engine := gin.Default()
	handler.InitHandlers(engine)
	err := engine.Run(":8080")
	if err != nil {
		log.WithError(err).Fatal("fail to serve with Gin")
	}
}
