package router

import (
	"log"

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
	// cfg := Config.GetGlobalConfig()
	if c.Request.Method == "GET" {
		c.HTML(200, "strawberry_400", gin.H{
			"message": "",
		})
	} else if c.Request.Method == "POST" {
		log.Fatalf("cannot post in page 400")
	}
}

func Error403(c *gin.Context) {
	// cfg := Config.GetGlobalConfig()
	if c.Request.Method == "GET" {
		c.HTML(200, "strawberry_403", gin.H{
			"message": "",
		})
	} else if c.Request.Method == "POST" {
		log.Fatalf("cannot post in page 403")
	}
}

func Error404(c *gin.Context) {
	// cfg := Config.GetGlobalConfig()
	if c.Request.Method == "GET" {
		c.HTML(200, "strawberry_404", gin.H{
			"message": "",
		})
	} else if c.Request.Method == "POST" {
		log.Fatalf("cannot post in page 404")
	}
}

func Error500(c *gin.Context) {
	// cfg := Config.GetGlobalConfig()
	if c.Request.Method == "GET" {
		c.HTML(200, "strawberry_500", gin.H{
			"message": "",
		})
	} else if c.Request.Method == "POST" {
		log.Fatalf("cannot post in page 500")
	}
}

func Error503(c *gin.Context) {
	// cfg := Config.GetGlobalConfig()
	if c.Request.Method == "GET" {
		c.HTML(200, "strawberry_503", gin.H{
			"message": "",
		})
	} else if c.Request.Method == "POST" {
		log.Fatalf("cannot post in page 503")
	}
}
