package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Mobile struct {
	a int
}

// 用户登录页面
func (m *Mobile) Dial(c *gin.Context) {
	if c.Request.Method == "GET" {
		c.HTML(http.StatusOK, "remark/static/mobile/mobile.html", gin.H{
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
