package handler

import (
	"CloudDrive/service"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"time"
)

type EmailForm struct {
	Email     string `form:"email" binding:"required,email"`
	EmailType string `form:"type" binding:"required"`
}

func RegisterEmailsRoutes(router *gin.Engine) {
	group := router.Group("/api/v1/emails")
	group.POST("", sendEmail)
	group.POST(":emailID", getEmailData)
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
		err = rdb.Set(ctx, emailID, code, configs.File.StaleTime).Err()
		if err != nil {
			c.JSON(500, gin.H{"message": "failed to store authentication code", "description": err.Error()})
			return
		}
		// return emailID
		c.JSON(200, gin.H{"emailID": emailID})
	} else {
		c.JSON(400, gin.H{"message": "input type is invalid"})
		return
	}
}

func getEmailData(c *gin.Context) {
	emailID := c.Param("emailID")
	log.WithFields(logrus.Fields{
		"emailID": emailID,
	}).Debug("get emailID from url")
	var emailForm EmailForm
	err := c.Bind(&emailForm)
	if err != nil {
		c.JSON(400, gin.H{"message": "invalid input data", "description": err.Error()})
		return
	}
	if emailForm.EmailType == "authCode" {
		code, err := rdb.Get(ctx, emailID).Result()
		if err != nil {
			c.JSON(404, gin.H{"message": "code for input emailID not found", "description": err.Error()})
			return
		}
		c.JSON(200, gin.H{"code": code})
	} else {
		c.JSON(400, gin.H{"message": "input type is invalid"})
		return
	}
}
