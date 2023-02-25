package handler

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func RegisterUsersRoutes(router *gin.Engine, sessionInfo *SessionInfo) {
	group := router.Group("/api/v1/users", sessions.Sessions(sessionInfo.Name, sessionInfo.Store))
	group.POST("", createUser)
}

func createUser(ctx *gin.Context) {

}
