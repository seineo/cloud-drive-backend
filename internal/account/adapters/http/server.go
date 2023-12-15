package http

import (
	"account/adapters/http/handler"
	applicationService "account/application/service"
	"account/config"
	"account/domain/account/entity"
	"common/middleware"
	"fmt"
	"github.com/gin-contrib/sessions"
	sessionRedis "github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"time"
)

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
		logrus.WithError(err).Fatal("fail to connect mysqlEventStore database")
	}

	// account 依赖注入
	accountFactoryConfig := entity.AccountFactoryConfig{
		NicknameRegex: "^[a-zA-Z_][a-zA-Z0-9_-]{0,38}$",
		PasswordRegex: "^[A-Za-z0-9]{6,38}$",
	}
	// code 依赖注入
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	codeFactory, err := entity.NewCodeFactory(5, time.Now().UnixNano())
	if err != nil {
		logrus.Fatal(err.Error())
	}

	applicationAccount := applicationService.NewApplicationAccount(accountFactoryConfig, codeFactory, mysqlDB, rdb)
	handler.RegisterAccountRoutes(hg.engine, applicationAccount)

	//运行http服务器
	err = hg.engine.Run()
	if err != nil {
		logrus.WithError(err).Fatal("unable to run handler server")
	}
}
