package controllers

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	Model "tomato/models"
)

//用于存储用户的切片
var Slice []Model.User
//用于临时存储用户登录信息的Map
var State = make(map[string]interface{})

// New is
func (u *Model.User) New() *Model.User {
	return &Model.User{
		ID:       primitive.NewObjectID(),
		Name:     u.Name,
		UserName: u.UserName,
		Email:    u.Email,
		Address:  u.Address,
		Phone:    u.Phone,
		Website:  u.Website,
		Company:  u.Company,
		Created:  time.Now(),
		Updated:  time.Now(),
	}
}

// 用户登录页面
func (a *Model.User) Login(c *gin.Context) {
	if c.Request.Method == "GET" {
			c.HTML(http.StatusOK, "remark/web/login/login.html", gin.H{
			"message": "",
		})
	} else if c.Request.Method == "POST" {
		username := c.Request.PostFormValue("Model.Username")
		password := c.Request.PostFormValue("password")
		// http.StatusOK == 200
		c.JSON(http.StatusOK, gin.H{
			//"hello": session.Get("mysession"),
			"Model.Username": Model.Username,
			"password": password,
		})
	}
}

// 验证用户是否登录
func (a *Model.User) IsLogin(c *gin.Context) {
}

// 添加用户
func (a *Model.User) AddUser(c *gin.Context) {
}

// 修改用户
func (a *Model.User) UpdateUser(c *gin.Context) {
}

// 删除用户
func (a *Model.User) DelUser(c *gin.Context) {
}


func (a *Model.User) LockScreen(c *gin.Context) {
	if c.Request.Method == "GET" {
			c.HTML(http.StatusOK, "remark/web/login/lockscreen.html", gin.H{
			"message": "",
		})
	} else if c.Request.Method == "POST" {
		username := c.Request.PostFormValue("Model.Username")
		password := c.Request.PostFormValue("password")
		c.JSON(http.StatusOK, gin.H{
			//"hello": session.Get("mysession"),
			"Model.Username": username,
			"password": password,
		})
	}
}

func (a *Model.User) Register(c *gin.Context) {
	if c.Request.Method == "GET" {
			c.HTML(http.StatusOK, "remark/web/login/register.html", gin.H{
			"message": "",
		})
	} else if c.Request.Method == "POST" {
		username := c.Request.PostFormValue("username")
		password := c.Request.PostFormValue("password")
		c.JSON(http.StatusOK, gin.H{
			//"hello": session.Get("mysession"),
			"username": username,
			"password": password,
		})
	}
}