package controllers

import (
	"github.com/gin-gonic/gin"
	// "tomato/controllers"
)

type Error struct {
	// The general error message
	//
	// required: true
	// example: Unauthorized
	Error string `json:"error"`
	// The http error code.
	//
	// required: true
	// example: 401
	ErrorCode int `json:"errorCode"`
	// The http error code.
	//
	// required: true
	// example: you need to provide a valid access token or user credentials to access this api
	ErrorDescription string `json:"errorDescription"`
}

// 用户登录页面
func Error400(c *gin.Context) {
	if c.Request.Method == "GET" {
			c.HTML(200, "remark/web/error/400.html", gin.H{
			"message": "",
		})
	} else if c.Request.Method == "POST" {

	}
}

func Error403(c *gin.Context) {
	if c.Request.Method == "GET" {
			c.HTML(200, "remark/web/error/403.html", gin.H{
			"message": "",
		})
	} else if c.Request.Method == "POST" {
		
	}
}

func Error404(c *gin.Context) {
	if c.Request.Method == "GET" {
			c.HTML(200, "remark/web/error/404.html", gin.H{
			"message": "",
		})
	} else if c.Request.Method == "POST" {
		
	}
}

func Error500(c *gin.Context) {
	if c.Request.Method == "GET" {
			c.HTML(200, "remark/web/error/500.html", gin.H{
			"message": "",
		})
	} else if c.Request.Method == "POST" {
		
	}
}

func Error503(c *gin.Context) {
	if c.Request.Method == "GET" {
			c.HTML(200, "remark/web/error/503.html", gin.H{
			"message": "",
		})
	} else if c.Request.Method == "POST" {
		
	}
}