package main

import (
	"account/adapters/http/handler"
	applicationService "account/application/service"
	"account/domain/account/entity"
	domainService "account/domain/account/service"
	"account/infrastructure/repo"
	"common/config"
	"common/logs"
	"common/middleware"
	"context"
	"fmt"
	"github.com/gin-contrib/sessions"
	sessionRedis "github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	redis "github.com/go-redis/redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"time"
)

//
//func init() {
//	log = config.GetLogger()
//	configs := config.LoadConfig([]string{"config"})
//	// create file storage directories if they don't exist
//	err := os.MkdirAll(configs.Local.StoragePath, 0750)
//	if err != nil {
//		log.Fatal("failed to create file storage directory")
//	}
//	err = os.MkdirAll(configs.Local.TempStoragePath, 0750)
//	if err != nil {
//		log.Fatal("failed to create temporal file storage directory")
//	}
//}

type HttpServer struct {
	config *config.Config
	engine *gin.Engine
}

func NewHttpServer(configs *config.Config, engine *gin.Engine) *HttpServer {
	return &HttpServer{config: configs, engine: engine}
}

func (hg *HttpServer) Run() {
	// 设置session中间件
	store, err := sessionRedis.NewStore(hg.config.RedisIdleConn, hg.config.RedisNetwork,
		hg.config.RedisAddr, hg.config.RedisPassword, []byte(hg.config.RedisKey))
	if err != nil {
		logrus.WithError(err).Fatal("fail to connect redis for middleware sessions")
	}
	// 设置全局中间件
	hg.engine.Use(sessions.Sessions("session_id", store), middleware.LoggingMiddleware())

	// 设置mysql
	dsn := fmt.Sprintf("%s:%s@%s(%s)/%s?parseTime=true",
		hg.config.DBUser, hg.config.DBPassword, hg.config.DBProtocol, hg.config.DBAddr, hg.config.DBDatabase)
	logrus.WithFields(logrus.Fields{
		"dsn": dsn,
	}).Debug(" dsn")

	mysqlDB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logrus.WithError(err).Fatal("fail to connect mysql database")
	}

	// account 依赖注入
	accountFactoryConfig := entity.AccountFactoryConfig{
		NicknameRegex: "^[a-zA-Z_][a-zA-Z0-9_-]{0,38}$",
		PasswordRegex: "^[A-Za-z0-9]{6,38}$",
	}
	accountRepo, err := repo.NewAccountRepo(mysqlDB)
	if err != nil {
		logrus.Fatal(err.Error())
	}
	// code 依赖注入
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	codeRepo, err := repo.NewCodeRepo(rdb, context.Background())
	if err != nil {
		logrus.Fatal(err.Error())
	}
	codeFactory, err := entity.NewCodeFactory(5, time.Now().UnixNano())
	if err != nil {
		logrus.Fatal(err.Error())
	}

	accountService := domainService.NewAccountService(accountRepo, accountFactoryConfig)
	verificationService := domainService.NewVerificationService(codeRepo, codeFactory)
	applicationAccount := applicationService.NewApplicationAccount(accountService, verificationService)
	handler.RegisterAccountRoutes(hg.engine, applicationAccount)

	// 运行
	err = hg.engine.Run()
	if err != nil {
		logrus.WithError(err).Fatal("failed to run handler server")
	}

}

func main() {
	logs.Init()
	// 读取配置
	configs, err := config.LoadConfig("../common/config")
	if err != nil {
		logrus.Fatal(err.Error())
	}
	// 运行http服务器
	engine := gin.Default()
	server := NewHttpServer(configs, engine)
	server.Run()
}
