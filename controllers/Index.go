package controllers

import (
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
	return func(c *gin.Context){
		// 获取客户端cookie并校验
		if cookie,err := c.Cookie("auth");err == nil{
			if cookie == "true"{ //校验是否有key为auth,value为true的cookie
				c.Next()
				return
			}
		}
		// 否则就返回无权限
		c.JSON(http.StatusUnauthorized,gin.H{"message":"操作非法"})
		// 如果验证不通过,不在调用后续的函数处理,直接从中间件就返回请求
		c.Abort()
		return
	}
}

// 用户登录页面
func IndexMain(c *gin.Context) {
	if c.Request.Method == "GET" {
			c.HTML(http.StatusOK, "remark/web/index/index.html", gin.H{
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