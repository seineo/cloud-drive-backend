package handler

import (
	"CloudDrive/config"
	"CloudDrive/service"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// SessionInfo info for middleware sessions
type SessionInfo struct {
	Name  string
	Store sessions.Store
}

var sessionInfo *SessionInfo
var log *logrus.Logger

func init() {
	log = service.GetLogger()

	redisConfig := config.GetConfig().Redis
	// specific where session stores and the key for authentication
	store, err := redis.NewStore(redisConfig.IdleConnection, redisConfig.Network,
		redisConfig.Addr, redisConfig.Password, []byte(redisConfig.AuthKey))
	if err != nil {
		log.WithError(err).Fatal("fail to connect redis for middleware sessions")
	}
	log.Info("connected redis for middleware sessions")
	sessionInfo = &SessionInfo{
		Name:  "session_id",
		Store: store,
	}
}

// InitHandlers initialize handlers with route groups and middlewares
func InitHandlers(router *gin.Engine) {

}
