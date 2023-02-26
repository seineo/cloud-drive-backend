package middleware

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func AuthCheck(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get("userID")
	if user == nil {
		msg := "User has not logged in, authentication failed"
		log.Info(msg)
		c.JSON(401, gin.H{"error": msg})
		c.Abort()
	}
}
