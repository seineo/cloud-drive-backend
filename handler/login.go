package handler

import (
	"CloudDrive/config"
	"CloudDrive/logic"
	"CloudDrive/model"
	"context"
	"github.com/alexedwards/argon2id"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v9"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

var ctx = context.Background()
var rdb *redis.Client

const expiredTime = 15 * time.Minute

func init() {
	// redis
	rdb = redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPsw,
		DB:       0, // use default DB
	})
}

// Get posted user data and check the validity,
// if success then store verification code, otherwise return err message
func register(c *gin.Context) {
	name := c.PostForm("name")
	email := c.PostForm("email")
	err := logic.CheckUser(name, email)
	if err != nil {
		log.WithFields(log.Fields{
			"err":      err.Error(),
			"username": name,
			"email":    email,
		}).Error("user info invalid")
		c.JSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
		return
	}
	code, err := logic.SendCodeEmail(email)
	if err != nil {
		log.WithFields(log.Fields{
			"email": email,
			"code":  code,
			"err":   err.Error(),
		}).Error("send verification code email error")
		return
	}
	err = rdb.Set(ctx, email, code, expiredTime).Err() // store email:code into redis with expiration time
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Error("failed to write verification code to redis")
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "fail to store verification code"})
	}
}

// Check posted verification code,
// if equals then store user data, otherwise return error message
func verify(c *gin.Context) {
	name := c.PostForm("name")
	email := c.PostForm("email")
	password := c.PostForm("password")
	inputCode := c.PostForm("code")
	storedCode, err := rdb.Get(ctx, email).Result()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "verification code for this email has expired"})
		return
	}
	if inputCode == storedCode { // match code, then store user data
		// hash password
		hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
		if err != nil {
			log.WithFields(log.Fields{
				"err":      err.Error(),
				"password": password,
			}).Error("fail to hash password using argon2id")
			c.JSON(http.StatusBadRequest, gin.H{"msg": "password hashing error"})
			return
		}
		// store user
		userID, err := model.AddUser(name, email, hash)
		if err != nil {
			log.WithFields(log.Fields{
				"err":  err.Error(),
				"user": userID,
			}).Error("fail to add user")
			c.JSON(http.StatusBadRequest, gin.H{"msg": "new user storage error"})
			return
		}
		// redirect to home page
		log.WithFields(log.Fields{
			"user": name,
		}).Info("user registered")
		session := sessions.Default(c)
		session.Set("user_id", userID)
		session.Save()
		c.Redirect(http.StatusMovedPermanently, "/")
	} else { // fail to match
		c.JSON(http.StatusBadRequest, gin.H{"msg": "verification code is incorrect"})
		return
	}
}
