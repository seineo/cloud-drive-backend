package handler

import (
	"CloudDrive/middleware"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func RegisterSessionsRoutes(router *gin.Engine) {
	group := router.Group("/api/v1/sessions")
	group.POST("", login)
	group.DELETE("current_session", middleware.AuthCheck, logout)
}

func login(c *gin.Context) {

}

func logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.JSON(200, gin.H{})
}
