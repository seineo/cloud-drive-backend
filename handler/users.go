package handler

import (
	"CloudDrive/model"
	"github.com/alexedwards/argon2id"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"time"
)

func RegisterUsersRoutes(router *gin.Engine) {
	group := router.Group("/api/v1/users")
	group.POST("", register)
}

func register(c *gin.Context) {
	var user model.User
	err := c.Bind(&user)
	if err != nil {
		c.JSON(400, gin.H{"message": "invalid input data", "description": err.Error()})
		return
	}
	// hash password using argon
	hash, err := argon2id.CreateHash(user.Password, argon2id.DefaultParams)
	user.Password = hash

	// check whether the email has been used
	_, err = model.GetUserByEmail(user.Email)
	if err == nil {
		c.JSON(409, gin.H{"message": "email has already been used"})
		return
	}

	// generate root dir hash for the user
	err = model.StoreFileMetadata(&model.File{
		Hash:       user.RootHash,
		Name:       "我的云盘",
		UserID:     user.ID,
		FileType:   "dir",
		Size:       0,
		DirPath:    "",
		Location:   "",
		CreateTime: time.Now(),
	})

	err = model.CreateUser(&user)
	if err != nil {
		c.JSON(500, gin.H{"message": "failed to create a user", "description": err.Error()})
		return
	}

	log.WithFields(logrus.Fields{
		"userID":    user.ID,
		"userName":  user.Name,
		"userEmail": user.Email,
	}).Info("created a new user")
	c.JSON(200, gin.H{"user": user})
}

func getUserInfo() {

}
