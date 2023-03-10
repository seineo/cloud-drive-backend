package main

import (
	"CloudDrive/handler"
	"CloudDrive/service"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"runtime"
)

var log *logrus.Logger

func init() {
	log = service.GetLogger()
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
