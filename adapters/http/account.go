package http

import (
	"CloudDrive/application/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

type AccountHandler struct {
	applicationAccount service.ApplicationAccount
}

func NewAccountHandler(applicationAccount service.ApplicationAccount) *AccountHandler {
	return &AccountHandler{applicationAccount: applicationAccount}
}

func registerAccountRoutes(router *gin.Engine, applicationAccount service.ApplicationAccount) {
	handler := NewAccountHandler(applicationAccount)

	group := router.Group("/api/v1/accounts")
	group.POST("", handler.register)
	group.PATCH("/me", updateAccount)
	group.POST("/sessions", login)
	group.DELETE("/sessions/me", logout)

}

func (ah *AccountHandler) register(c *gin.Context) {
	var user UserSignUpRequest
	if err := c.ShouldBindJSON(&user); err != nil {
		InvalidInputErr(c, err, "invalid user signup request data")
		return
	}
	account, err := ah.applicationAccount.Create(user)
	if err != nil {
		RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, account)
}

func updateAccount(c *gin.Context) {

}

func login(c *gin.Context) {

}

func logout(c *gin.Context) {

}
