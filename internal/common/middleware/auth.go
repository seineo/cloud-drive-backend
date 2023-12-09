package middleware

import (
	"common/server"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		accountID := session.Get("id")
		if accountID == nil {
			server.UnauthorizedErr(c, fmt.Errorf("account id not found in session keys"), "user has not logged in")
			c.Abort()
		}
	}
}
