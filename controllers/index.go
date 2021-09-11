package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

//1. 有俩个路由，login和home
//2. login用于设置cookie
//3. home是访问查看信息的请求
//4. 在请求home之前，先跑中间件代码，检验是否存在cookie
//5. 如果没有login设置cookie，就直接访问home，会显示无权限，因为权限校验没有通过
// 权限校验中间件
func AuthMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取客户端cookie并校验
		if cookie, err := c.Cookie("auth"); err == nil {
			if cookie == "true" { //校验是否有key为auth,value为true的cookie
				c.Next()
				return
			}
		}
		// here you can add your authentication method to authorize users.
		username := c.PostForm("user")
		password := c.PostForm("password")
		fmt.Println("username:" + username + " " + "password: " + password)

		// if username == "foo" && password == "bar" {
		// 	return
		// } else {
		// 	c.AbortWithStatus(http.StatusUnauthorized)
		// }

		// 否则就返回无权限
		c.JSON(http.StatusUnauthorized, gin.H{"message": "operation not allowed"})
		// 如果验证不通过,不在调用后续的函数处理,直接从中间件就返回请求
		c.Abort()
		return
	}
}

type StrawberryContext struct {
	Router *gin.Engine
}

var StawCtx StrawberryContext

func SetRouter(router *gin.Engine) {
	StawCtx.Router = router
}

func RedirectLogin(c *gin.Context) {
	c.Request.URL.Path = "/login"
	// 路由重定向, 使用 HandleContext
	StawCtx.Router.HandleContext(c)

	// 	c.Redirect(http.StatusMovedPermanently, "http://www.google.com/")
	// c.Redirect(http.StatusFound, "/foo")
}

// 用户登录页面
func IndexMain(c *gin.Context) {
	if c.Request.Method == "GET" {
		RedirectLogin(c)
		// c.HTML(http.StatusOK, "remark/static/index/index.html", gin.H{
		// 	"title":   "Dashboard",
		// 	"message": "",
		// })
	} else if c.Request.Method == "POST" {
		// username := c.Request.PostFormValue("username")
		// password := c.Request.PostFormValue("password")
		// http.StatusOK == 200
		// c.JSON(http.StatusOK, gin.H{
		// 	// "hello":    session.Get("mysession"),
		// 	"username": username,
		// 	"password": password,
		// })
		c.HTML(http.StatusOK, "remark/static/index/index.html", gin.H{
			"title":   "Dashboard",
			"message": "",
		})
	}
}
