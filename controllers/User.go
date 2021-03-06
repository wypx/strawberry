package controllers

import (
	"fmt"
	"time"
	"net/http"
	"github.com/gin-gonic/gin"
	Model "tomato/models"
	DataBase "tomato/database"
)

type User Model.User 
type Response Model.Response 

//用于存储用户的切片
var Slice []User
//用于临时存储用户登录信息的Map
var State = make(map[string]interface{})

// New is
func (u *User) New() *User {
	return &User{
		UserName: u.UserName,
		PassWord: u.PassWord,
		Email:    u.Email,
		Address:  u.Address,
		Phone:    u.Phone,
		WebSite:  u.WebSite,
		Created:  time.Now(),
		Updated:  time.Now(),
	}
}

// 用户登录页面
func (a *User) Login(c *gin.Context) {
	if c.Request.Method == "GET" {
			c.HTML(http.StatusOK, "remark/web/login/login.html", gin.H{
			"message": "",
		})
	} else if c.Request.Method == "POST" {
		username := c.Request.PostFormValue("username")
		password := c.Request.PostFormValue("password")

		if len(username) == 0 || len(password) == 0 {
			rsp := Response{
						Code:    50001,
						Status:  "error",
						Message: "账号或密码不能为空",
						Data:    "",
				}
				c.JSON(http.StatusOK, rsp)
		} else {
				var user User
				DataBase.GetDB().Where("user_name=?", username).First(&user)
				if user.UserName == "" {
					rsp := Response{
							Code:    50001,
							Status:  "error",
							Message: "用户不存在",
							Data:    "",
					}
					c.JSON(http.StatusOK, rsp)
				} else {
					if user.PassWord == password {
						fmt.Println("Login successful")
						c.Redirect(http.StatusMovedPermanently, "/")
					} else {
						rsp := Response{
									Code:    5001,
									Status:  "error",
									Message: "密码错误",
									Data:    "",
							}
							// c.BindJSON(&user)
							// c.JSON(200, user)
							c.JSON(http.StatusOK, rsp)
					}
				}
		}
	}
}

// 验证用户是否登录
func IsLogin(c *gin.Context) bool {
	return true
}

// 添加用户
func AddUser(user *User) bool {
	DataBase.GetDB().Create(user)
	return true
}

// 修改用户
func UpdateUser(user *User, pass string) bool {
	DataBase.GetDB().Model(user).Update("password", pass)
	return true
}

// 删除用户
func DelUser(user *User) bool {
  DataBase.GetDB().Delete(user)
	return true
}


func (a *User) LockSess(c *gin.Context) {
	if c.Request.Method == "GET" {
			c.HTML(http.StatusOK, "remark/web/login/lockscreen.html", gin.H{
			"message": "",
		})
	} else if c.Request.Method == "POST" {
		username := c.Request.PostFormValue("Username")
		password := c.Request.PostFormValue("password")
		c.JSON(http.StatusOK, gin.H{
			//"hello": session.Get("mysession"),
			"Username": username,
			"password": password,
		})
	}
}

func (a *User) Register(c *gin.Context) {
	if c.Request.Method == "GET" {
			c.HTML(http.StatusOK, "remark/web/login/register.html", gin.H{
			"message": "",
		})
	} else if c.Request.Method == "POST" {
		username := c.Request.PostFormValue("username")
		password := c.Request.PostFormValue("password")
		if len(username) == 0 || len(password) == 0 {
			rsp := Response{
						Code:    50001,
						Status:  "error",
						Message: "账号或密码不能为空",
						Data:    "",
				}
				c.JSON(http.StatusOK, rsp)
		} else {
				var user User
				DataBase.GetDB().Where("user_name=?", username).First(&user)
				if user.UserName == username {
					rsp := Response{
							Code:    50001,
							Status:  "error",
							Message: "用户已注册",
							Data:    "",
					}
					c.JSON(http.StatusOK, rsp)
				} else {
					newUser := User{UserName: username, PassWord: password}
					AddUser(&newUser)
					// https://www.cnblogs.com/yh2924/p/12383317.html
					// https://www.cnblogs.com/wangyuyu/p/12023270.html
					c.Redirect(http.StatusMovedPermanently, "/login")
				}
		}
	}
}

func GetUserInfo(c *gin.Context) {
	id := c.Query("ID")
	if id == "" {
			rsp := Response{
					Code:    50001,
					Status:  "error",
					Message: "参数错误",
					Data:    "",
			}
			c.JSON(200, rsp)
			// c.AbortWithStatus(404)
			return
	}
	var user User
	DataBase.GetDB().First(&user, id)
	if user.ID > 0 {
			rsp := Response{
					Code:    200,
					Status:  "success",
					Message: "",
					Data:    user,
			}
			c.JSON(200, rsp)
	} else {
			rsp := Response{
					Code:    50001,
					Status:  "error",
					Message: "用户不存在",
					Data:    "",
			}
			c.JSON(200, rsp)
	}
}

func GetAllUserInfo(c *gin.Context) {
	var users []User
	if err := DataBase.GetDB().Find(&users).Error; err != nil {
		 c.AbortWithStatus(404)
		 fmt.Println(err)
	} else {
		 c.JSON(200, users)
	}
}