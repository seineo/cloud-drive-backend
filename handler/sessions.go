package handler

import (
	"CloudDrive/middleware"
	"CloudDrive/model"
	"CloudDrive/response"
	"github.com/alexedwards/argon2id"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type UserLogin struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func RegisterSessionsRoutes(router *gin.Engine) {
	group := router.Group("/api/v1/sessions")
	group.POST("", login)

	group.DELETE("current_session", middleware.AuthCheck, logout)
}

func login(c *gin.Context) {
	var userLogin UserLogin
	err := c.Bind(&userLogin)
	if err != nil {
		c.JSON(400, gin.H{"message": "invalid input data", "description": err.Error()})
		return
	}

	user, err := model.GetUserByEmail(userLogin.Email)
	if err != nil {
		c.JSON(400, gin.H{"message": "email not found"})
		return
	}
	log.WithFields(logrus.Fields{
		"input":  userLogin.Password,
		"hashed": user.Password,
	}).Debug("compare password")
	match, err := argon2id.ComparePasswordAndHash(userLogin.Password, user.Password)
	if err != nil {
		c.JSON(500, gin.H{"message": "failed to compare password", "description": err.Error()})
		return
	}
	if match {
		session := sessions.Default(c)
		session.Set("userID", user.ID)
		err = session.Save()
		if err != nil {
			c.JSON(500, gin.H{"message": "failed to save session", "description": err.Error()})
			return
		}
		log.WithFields(logrus.Fields{
			"userID":    user.ID,
			"userName":  user.Name,
			"userEmail": user.Email,
		}).Info("user logged in")
	} else {
		c.JSON(400, gin.H{"message": "wrong password"})
		return
	}
	c.JSON(200, response.UserSignResponse{
		Email:    user.Email,
		Name:     user.Name,
		RootHash: user.RootHash,
	})
}

func logout(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("userID")
	session.Clear() // this will mark the session as "written" only if there's at least one key to delete
	session.Options(sessions.Options{MaxAge: -1})
	err := session.Save()
	if err != nil {
		c.JSON(500, gin.H{"message": "failed to logout", "description": err.Error()})
	}
	log.WithFields(logrus.Fields{
		"userID": userID,
	}).Info("user logged out")
	c.Status(204)
}
