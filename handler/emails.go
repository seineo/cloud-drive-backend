package handler

import (
	"github.com/gin-gonic/gin"
)

func RegisterEmailsRoutes(router *gin.Engine) {
	group := router.Group("/api/v1/emails")
	group.POST("", sendEmail)
	group.GET(":email_id", getEmailData)
}

func sendEmail(c *gin.Context) {

}

func getEmailData(c *gin.Context) {

}
