package router

import (
	"fmt"
	"log"
	"net/http"
	"time"
	DataBase "vm_manager/vm_admin/database"
	Model "vm_manager/vm_admin/models"

	"github.com/gin-gonic/gin"
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
	// cfg := Config.GetGlobalConfig()
	if c.Request.Method == "GET" {
		StawCtx.Set("gin-login", false)
		log.Printf("=====================login page========================")
		c.HTML(http.StatusOK, "strawberry_login", gin.H{
			"strawberry_title": "登录",
		})
	} else if c.Request.Method == "POST" {
		username := c.Request.PostFormValue("username")
		password := c.Request.PostFormValue("password")

		log.Printf("user: " + username + " pass: " + password)

		if len(username) == 0 || len(password) == 0 {
			// rsp := Response{
			// 	Code:    50001,
			// 	Status:  "error",
			// 	Message: "user or pass cannot be empty",
			// 	Data:    "",
			// }
			// c.JSON(http.StatusOK, rsp)
			c.HTML(http.StatusOK, "strawberry_login", gin.H{
				"strawberry_title":  "登录",
				"strawberry_status": "user or pass cannot be empty",
			})
		} else {
			var user User
			DataBase.GetDB().Where("user_name=?", username).First(&user)
			if user.UserName == "" {
				rsp := Response{
					Code:    50001,
					Status:  "error",
					Message: "user not exist",
					Data:    "",
				}
				c.JSON(http.StatusOK, rsp)
			} else {
				if user.PassWord == password {
					fmt.Printf("login successful")
					StawCtx.Set("gin-login", true)
					c.Redirect(http.StatusMovedPermanently, "/")
				} else {
					rsp := Response{
						Code:    5001,
						Status:  "error",
						Message: "pass is incorrect",
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
		c.HTML(http.StatusOK, "strawberry_lockscreen", gin.H{
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
		c.HTML(http.StatusOK, "strawberry_register", gin.H{
			"message": "",
		})
	} else if c.Request.Method == "POST" {
		username := c.Request.PostFormValue("username")
		password := c.Request.PostFormValue("password")
		if len(username) == 0 || len(password) == 0 {
			rsp := Response{
				Code:    50001,
				Status:  "error",
				Message: "user or pass cannot be empty",
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
					Message: "user registered yet",
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
			Message: "parameter error",
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
			Message: "user not exist",
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
