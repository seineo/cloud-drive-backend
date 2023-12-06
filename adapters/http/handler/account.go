package handler

import (
	"CloudDrive/adapters/http/types"
	"CloudDrive/application/service"
	"CloudDrive/domain/account/entity"
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
	group.DELETE("/sessions/me", handler.logout)
	group.PATCH("/me", handler.updateAccount)
	group.DELETE("/me", handler.deleteAccount)
}

func setSession(c *gin.Context, account *entity.Account) {
	session := sessions.Default(c)
	session.Set("id", account.GetID())
	session.Set("email", account.GetEmail())
	session.Set("nickname", account.GetNickname())
	err := session.Save()
	if err != nil {
		RespondWithError(c, err)
		return
	}
}

func clearSession(c *gin.Context, session sessions.Session) {
	session.Clear()
	session.Options(sessions.Options{MaxAge: -1})
	err := session.Save()
	if err != nil {
		RespondWithError(c, err)
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
		InvalidInputErr(c, err, "invalid request data for user registration")
		return
	}
	account, err := ah.applicationAccount.Create(user)
	if err != nil {
		RespondWithError(c, err)
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
		InvalidInputErr(c, err, "invalid request data for user login")
		return
	}
	account, err := ah.applicationAccount.Login(user)
	if err != nil {
		RespondWithError(c, err)
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

func (ah *AccountHandler) updateAccount(c *gin.Context) {
	var user types.AccountUpdateRequest
	if err := c.ShouldBindJSON(&user); err != nil {
		InvalidInputErr(c, err, "invalid request data for user profile update")
		return
	}
	session := sessions.Default(c)
	accountID := session.Get("id")
	if err := ah.applicationAccount.Update(accountID.(uint), user); err != nil {
		RespondWithError(c, err)
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
		RespondWithError(c, err)
		return
	}
	logrus.WithFields(logrus.Fields{"id": accountID}).Info("user delete account")
	c.Status(http.StatusNoContent)
}
