package routes

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"net/http"
	"tomato/controllers"
)

func InitRouter() *gin.Engine {
	//使用gin的Default方法创建一个路由handler
	router := gin.Default()
	//设置默认路由当访问一个错误网站时返回
	router.NoRoute(controllers.Error404)

	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("mysession", store))

	router.StaticFS("/webroot", http.Dir("remark")) // 静态文件路径
	// https://gin-gonic.com/zh-cn/docs/examples/html-rendering/
	router.LoadHTMLGlob("remark/web/**/*.html") //渲染html页面

	index := router.Group("/")
	{
		index.GET("/", controllers.IndexMain)
		index.POST("/", controllers.IndexMain)
	}

	login := router.Group("/login")
	{
		User := new(controllers.User)
		login.GET("/", User.Login)
		login.POST("/", User.Login)
	}

	lock := router.Group("/lock")
	{
		User := new(controllers.User)
		lock.GET("/", User.LockScreen)
		lock.POST("/", User.LockScreen)
	}

	register := router.Group("/register")
	{
		User := new(controllers.User)
		register.GET("/", User.Register)
		register.POST("/", User.Register)
	}

	forgot := router.Group("/forgot")
	{
		forgot.GET("/", controllers.ForgotPass)
		forgot.POST("/", controllers.ForgotPass)
	}

	maintain := router.Group("/maintain")
	{
		maintain.GET("/", controllers.Maintain)
		maintain.POST("/", controllers.Maintain)
	}

	error400 := router.Group("/400")
	{
		error400.GET("/", controllers.Error400)
		error400.POST("/", controllers.Error400)
	}

	error403 := router.Group("/403")
	{
		error403.GET("/", controllers.Error403)
		error403.POST("/", controllers.Error403)
	}

	error404 := router.Group("/404")
	{
		error404.GET("/", controllers.Error404)
		error404.POST("/", controllers.Error404)
	}

	error500 := router.Group("/500")
	{
		error500.GET("/", controllers.Error500)
		error500.POST("/", controllers.Error500)
	}

	error503 := router.Group("/503")
	{
		error503.GET("/", controllers.Error503)
		error503.POST("/", controllers.Error503)
	}

	mobile := router.Group("/mobile")
	{
		Mobile := new(controllers.Mobile)
		mobile.GET("/", Mobile.Dial)
		mobile.POST("/", Mobile.Dial)
	}

	user := router.Group("/user")
	{
		User := new(controllers.User)
		user.GET("/login", User.Login)
		user.POST("/login", User.Login)
	}

	web := router.Group("/test")
	{
		adminUser := new(controllers.AdminUser)
		web.GET("/test", adminUser.Test)
		web.GET("/session_test", adminUser.SessionTest)
		web.GET("/template_test", adminUser.Template)
	}

	v1 := router.Group("/api/v1")
	{
		adminUser := new(controllers.AdminUser)
		v1.GET("/admin-users", adminUser.Index)
		v1.POST("/admin-users", adminUser.Store)
		v1.PATCH("/admin-users/:id", adminUser.Update)
		v1.DELETE("/admin-users/:id", adminUser.Destroy)
		v1.GET("/admin-users/:id", adminUser.Show)
	}
	return router
}
