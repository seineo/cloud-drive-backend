package handler

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
)

// GetRootPage returns login page if user has not logged in,
// otherwise return home page
func GetRootPage(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get("user_id")
	if user == nil { // not logged in
		c.HTML(
			http.StatusOK,
			"login.html",
			gin.H{})
	} else {
		c.HTML(
			http.StatusOK,
			"index.html",
			gin.H{
				"title": "HomePage",
			})
	}
}

// PostRootPage deal with register post, login post and other posts from home page
func PostRootPage(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	if userID == nil { // not logged in or registered
		name := c.PostForm("name")
		if name != "" { // register post
			code := c.PostForm("code")
			if code != "" { // verification
				verify(c)
			} else { // register
				register(c)
			}
		} else { // login post
			//login()
		}
	} else {

	}
}
