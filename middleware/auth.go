package middleware

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func AuthCheck(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get("id")
	if user == nil {
		log.Info("User has not logged in, Redirecting to login page...")
		c.Redirect(http.StatusMovedPermanently, "/")
		c.Abort()
		return
	}
}
