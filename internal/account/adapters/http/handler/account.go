package handler

import (
	"account/adapters/http/types"
	"account/application/service"
	"account/domain/account/entity"
	"common/middleware"
	"common/server"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
)

type AccountHandler struct {
	applicationAccount service.ApplicationAccount
}

func NewAccountHandler(applicationAccount service.ApplicationAccount) *AccountHandler {
	return &AccountHandler{applicationAccount: applicationAccount}
}

func RegisterAccountRoutes(router *gin.Engine, applicationAccount service.ApplicationAccount) {
	handler := NewAccountHandler(applicationAccount)

	group := router.Group("/api/v1/accounts")
	group.POST("", handler.register)
	group.POST("/sessions", handler.login)
	group.DELETE("/sessions/me", middleware.AuthMiddleware(), handler.logout)
	group.GET("/me", middleware.AuthMiddleware(), handler.getAccount)
	group.PATCH("/me", middleware.AuthMiddleware(), handler.updateAccount)
	group.DELETE("/me", middleware.AuthMiddleware(), handler.deleteAccount)
	group.POST("/codes", handler.sendVerificationCode)
	group.GET("/codes/:email", handler.getVerificationCode)
}

func setSession(c *gin.Context, account *entity.Account) {
	session := sessions.Default(c)
	session.Set("id", account.GetID())
	session.Set("email", account.GetEmail())
	session.Set("nickname", account.GetNickname())
	err := session.Save()
	if err != nil {
		server.RespondWithError(c, err)
		return
	}
}

func clearSession(c *gin.Context, session sessions.Session) {
	session.Clear()
	session.Options(sessions.Options{MaxAge: -1})
	err := session.Save()
	if err != nil {
		server.RespondWithError(c, err)
		return
	}
}

func domainToResponse(account *entity.Account) types.AccountResponse {
	return types.AccountResponse{
		Id:       account.GetID(),
		Email:    account.GetEmail(),
		Nickname: account.GetNickname(),
	}
}

func (ah *AccountHandler) register(c *gin.Context) {
	var user types.AccountSignUpRequest
	if err := c.ShouldBindJSON(&user); err != nil {
		server.InvalidInputErr(c, err, "invalid request data for user registration")
		return
	}
	account, err := ah.applicationAccount.Create(user)
	if err != nil {
		server.RespondWithError(c, err)
		return
	}
	logrus.WithFields(logrus.Fields{
		"id":       account.GetID(),
		"email":    user.Email,
		"nickname": user.Nickname}).Info("user register")
	c.JSON(http.StatusOK, domainToResponse(account))
}

func (ah *AccountHandler) login(c *gin.Context) {
	var user types.AccountLoginRequest
	if err := c.ShouldBindJSON(&user); err != nil {
		server.InvalidInputErr(c, err, "invalid request data for user login")
		return
	}
	account, err := ah.applicationAccount.Login(user)
	if err != nil {
		server.RespondWithError(c, err)
		return
	}
	// 存储session
	setSession(c, account)
	logrus.WithFields(logrus.Fields{"id": account.GetID()}).Info("user login")
	c.JSON(http.StatusOK, domainToResponse(account))
}

func (ah *AccountHandler) logout(c *gin.Context) {
	session := sessions.Default(c)
	accountID := session.Get("id")
	// 清除session
	clearSession(c, session)
	logrus.WithFields(logrus.Fields{"id": accountID}).Info("user logout")
	c.Status(http.StatusNoContent)
}

func (ah *AccountHandler) getAccount(c *gin.Context) {
	session := sessions.Default(c)
	accountID := session.Get("id")
	account, err := ah.applicationAccount.Get(accountID.(uint))
	if err != nil {
		server.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, domainToResponse(account))
}

func (ah *AccountHandler) updateAccount(c *gin.Context) {
	var user types.AccountUpdateRequest
	if err := c.ShouldBindJSON(&user); err != nil {
		server.InvalidInputErr(c, err, "invalid request data for user profile update")
		return
	}
	session := sessions.Default(c)
	accountID := session.Get("id")
	if err := ah.applicationAccount.Update(accountID.(uint), user); err != nil {
		server.RespondWithError(c, err)
		return
	}
	logrus.WithFields(logrus.Fields{"id": accountID}).Info("user update profile")
	c.Status(http.StatusNoContent)
}

func (ah *AccountHandler) deleteAccount(c *gin.Context) {
	session := sessions.Default(c)
	accountID := session.Get("id")
	// 清除session
	clearSession(c, session)
	// 删除账号
	if err := ah.applicationAccount.Delete(accountID.(uint)); err != nil {
		server.RespondWithError(c, err)
		return
	}
	logrus.WithFields(logrus.Fields{"id": accountID}).Info("user delete account")
	c.Status(http.StatusNoContent)
}

func (ah *AccountHandler) sendVerificationCode(c *gin.Context) {
	var user types.AccountCodeRequest
	if err := c.ShouldBindJSON(&user); err != nil {
		server.InvalidInputErr(c, err, "invalid request data for sending verification code")
		return
	}
	code, err := ah.applicationAccount.SendVerificationCode(user)
	if err != nil {
		server.RespondWithError(c, err)
		return
	}
	logrus.WithFields(logrus.Fields{"code": code, "email": user.Email}).Info("send verification code")
	c.JSON(http.StatusOK, code)
}

func (ah *AccountHandler) getVerificationCode(c *gin.Context) {
	email := c.Param("email")
	if len(email) == 0 {
		server.InvalidInputErr(c, fmt.Errorf("invalid email"), "email should not be empty")
	}
	code, err := ah.applicationAccount.GetVerificationCode(email)
	if err != nil {
		server.RespondWithError(c, err)
		return
	}
	logrus.WithFields(logrus.Fields{"code": code, "email": email}).Info("get verification code")
	c.JSON(http.StatusOK, code)
}
