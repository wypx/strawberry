package controllers

import (
	"time"
	"net/http"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"github.com/gin-gonic/gin"
)

type User struct {
	Basic
	ID       primitive.ObjectID `bson:"_id" json:"id"`
	Name     string             `bson:"name" json:"name"`
	UserName string             `bson:"username" json:"username"`
	Email    string             `bson:"email" json:"email"`
	Address  UserAddress        `bson:"address" json:"address"`
	Phone    string             `bson:"phone" json:"phone"`
	Website  string             `bson:"website" json:"website"`
	Company  UserCompany        `bson:"company" json:"company"`
	Created  time.Time          `bson:"created" json:"created"`
	Updated  time.Time          `bson:"updated" json:"updated"`
}


// The UserAddress holds
type UserAddress struct {
	Street  string         `bson:"street" json:"street"`
	Suite   string         `bson:"suite" json:"suite"`
	City    string         `bson:"city" json:"city"`
	Zipcode string         `bson:"zipcode" json:"zipcode"`
	Geo     UserAddressGeo `bson:"geo" json:"geo"`
}

// The UserAddressGeo holds
type UserAddressGeo struct {
	Lat string `bson:"lat" json:"lat"`
	Lng string `bson:"lng" json:"lng"`
}

// The UserCompany holds
type UserCompany struct {
	Name        string `bson:"name" json:"name"`
	CatchPhrase string `bson:"catchPhrase" json:"catchPhrase"`
	BS          string `bson:"bs" json:"bs"`
}

// New is
func (u *User) New() *User {
	return &User{
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
func (a *User) Login(c *gin.Context) {
	if c.Request.Method == "GET" {
			c.HTML(http.StatusOK, "remark/web/login/login.html", gin.H{
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

// 验证用户是否登录
func (a *User) IsLogin(c *gin.Context) {
}

// 添加用户
func (a *User) AddUser(c *gin.Context) {
}

// 修改用户
func (a *User) UpdateUser(c *gin.Context) {
}

// 删除用户
func (a *User) DelUser(c *gin.Context) {
}


func (a *User) LockScreen(c *gin.Context) {
	if c.Request.Method == "GET" {
			c.HTML(http.StatusOK, "remark/web/login/lockscreen.html", gin.H{
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

func (a *User) Register(c *gin.Context) {
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