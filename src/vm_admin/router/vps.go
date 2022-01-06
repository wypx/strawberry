package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 用户登录页面
func PageVPS(c *gin.Context) {
	if c.Request.Method == "GET" {
		c.HTML(http.StatusOK, "strawberry_vps", gin.H{
			"strawberry_title": "云主机",
		})
	} else if c.Request.Method == "POST" {
		c.HTML(http.StatusOK, "strawberry_vps", gin.H{
			"strawberry_title": "云主机",
		})
	}
}
