package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 用户登录页面
func Maintain(c *gin.Context) {
	if c.Request.Method == "GET" {
		c.HTML(http.StatusOK, "strawberry_maintain", gin.H{
			"message": "",
		})
	} else if c.Request.Method == "POST" {
		username := c.Request.PostFormValue("username")
		password := c.Request.PostFormValue("password")
		// http.StatusOK == 200
		c.JSON(http.StatusOK, gin.H{
			//"hello": session.Get("mysession"),
			"username": username,
			"password": password,
		})
	}
}
