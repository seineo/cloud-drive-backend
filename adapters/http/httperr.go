package http

import (
	"CloudDrive/common/slugerror"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
)

// 这几个错误是主动调用

func InvalidInputErr(c *gin.Context, err error, slug string) {
	httpRespondWithError(c, http.StatusBadRequest, err, slug)
}

func UnauthorizedErr(c *gin.Context, err error, slug string) {
	httpRespondWithError(c, http.StatusUnauthorized, err, slug)
}

func ConflictErr(c *gin.Context, err error, slug string) {
	httpRespondWithError(c, http.StatusConflict, err, slug)
}

func InternalServerErr(c *gin.Context, err error, slug string) {
	httpRespondWithError(c, http.StatusInternalServerError, err, slug)

}

// RespondWithError 让该函数来分析是什么错误并返回
func RespondWithError(c *gin.Context, err error) {
	slugErr, ok := err.(*slugerror.SlugError)
	if !ok {
		// 有些简单的、自己未定义SlugError的internal error
		InternalServerErr(c, err, "internal server error")
	}
	switch slugErr.ErrorType() { // 除了500错误，其他都应该自己定义SlugError， 让接入层知道返回什么状态码
	case slugerror.ErrInvalidInput:
		InvalidInputErr(c, slugErr, slugErr.Slug())
	case slugerror.ErrUnauthorized:
		UnauthorizedErr(c, slugErr, slugErr.Slug())
	case slugerror.ErrConflict:
		ConflictErr(c, slugErr, slugErr.Slug())
	default:
		InternalServerErr(c, slugErr, slugErr.Slug())
	}
}

func httpRespondWithError(c *gin.Context, code int, error error, slug string) {
	logrus.WithError(error).Error(slug)
	c.JSON(code, gin.H{"error": error.Error(), "message": slug})
}
