package handler

import (
	"CloudDrive/config"
	"CloudDrive/middleware"
	"CloudDrive/service"
	"context"
	"github.com/gin-contrib/sessions"
	sessionRedis "github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v9"
	"github.com/sirupsen/logrus"
)

// SessionInfo info for middleware sessions
type SessionInfo struct {
	Name  string
	Store sessions.Store
}

var sessionInfo *SessionInfo
var log *logrus.Logger
var ctx = context.Background()
var rdb *redis.Client

func init() {
	log = service.GetLogger()
	redisConfig := config.GetConfig().Redis
	// specific where session stores and the key for authentication
	store, err := sessionRedis.NewStore(redisConfig.IdleConnection, redisConfig.Network,
		redisConfig.Addr, redisConfig.Password, []byte(redisConfig.AuthKey))
	if err != nil {
		log.WithError(err).Fatal("fail to connect redis for middleware sessions")
	}
	log.Info("connected redis for middleware sessions")
	sessionInfo = &SessionInfo{
		Name:  "session_id",
		Store: store,
	}
	// redis client
	rdb = redis.NewClient(&redis.Options{
		Addr:     redisConfig.Addr,
		Password: redisConfig.Password, // no password set
		DB:       redisConfig.DB,       // use default DB
	})
}

// InitHandlers initialize handlers with route groups and middlewares
func InitHandlers(router *gin.Engine) {
	router.Use(sessions.Sessions(sessionInfo.Name, sessionInfo.Store), middleware.CORSMiddleware())
	RegisterUsersRoutes(router)
	RegisterSessionsRoutes(router)
	RegisterEmailsRoutes(router)
}
