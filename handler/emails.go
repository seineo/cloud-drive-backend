package handler

import (
	"CloudDrive/config"
	"CloudDrive/service"
	"github.com/gin-gonic/gin"
	"time"
)

type EmailForm struct {
	Email     string `form:"email" binding:"required,email"`
	EmailType string `form:"type" binding:"required"`
}

func RegisterEmailsRoutes(router *gin.Engine) {
	group := router.Group("/api/v1/emails")
	group.POST("", sendEmail)
	group.GET(":emailID", getEmailData)
}

func sendEmail(c *gin.Context) {
	var emailForm EmailForm
	err := c.Bind(&emailForm)
	if err != nil {
		c.JSON(400, gin.H{"message": "invalid input data", "description": err.Error()})
		return
	}
	if emailForm.EmailType == "authCode" {
		code, err := service.SendCodeEmail(emailForm.Email)
		if err != nil {
			c.JSON(500, gin.H{"message": "failed to send authentication email", "description": err.Error()})
			return
		}
		// store authentication code using pattern like emailID:code
		emailID := service.SHA256Hash(emailForm.Email, time.Now().String())
		err = rdb.Set(ctx, emailID, code, config.GetConfig().AuthCodeExpiredTime).Err()
		if err != nil {
			c.JSON(500, gin.H{"message": "failed to store authentication code", "description": err.Error()})
			return
		}
	} else {
		c.JSON(400, gin.H{"message": "input type is invalid"})
		return
	}
}

func getEmailData(c *gin.Context) {

}
