package handler

import (
	"CloudDrive/model"
	"github.com/gin-gonic/gin"
)

func RegisterUsersRoutes(router *gin.Engine) {
	group := router.Group("/api/v1/users")
	group.POST("", createUser)
}

func createUser(c *gin.Context) {
	var user model.User
	err := c.Bind(&user)
	if err != nil {
		c.JSON(400, gin.H{"message": "invalid input data", "description": err.Error()})
	}
	userID, err := model.CreateUser(&user)
	if err != nil {
		c.JSON(500, gin.H{"message": "failed to create a user", "description": err.Error()})
	}
	c.JSON(200, gin.H{"userID": userID})
}
