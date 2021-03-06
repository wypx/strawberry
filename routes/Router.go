package routes

import (
	// "os"
	// "io"
	"fmt"
	// "log"
	"time"
	"context"
	"net/http"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	Controller "tomato/controllers"
)
// timeout middleware wraps the request context with a timeout
func TimeoutMiddleware(timeout time.Duration) func(c *gin.Context) {
	return func(c *gin.Context) {
 
		// wrap the request context with a timeout
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
 
		defer func() {
			// check if context timeout was reached
			if ctx.Err() == context.DeadlineExceeded {
 
				// write response and abort the request
				c.Writer.WriteHeader(http.StatusGatewayTimeout)
				c.Abort()
			}
 
			//cancel to clear resources after finished
			cancel()
		}()
 
		// replace request with context wrapped request
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
 
func TimedHandler(duration time.Duration) func(c *gin.Context) {
	return func(c *gin.Context) {
 
		// get the underlying request context
		ctx := c.Request.Context()
 
		// create the response data type to use as a channel type
		type responseData struct {
			status int
			body   map[string]interface{}
		}
 
		// create a done channel to tell the request it's done
		doneChan := make(chan responseData)
 
		// here you put the actual work needed for the request
		// and then send the doneChan with the status and body
		// to finish the request by writing the response
		go func() {
			time.Sleep(duration)
			doneChan <- responseData{
				status: 200,
				body:   gin.H{"hello": "world"},
			}
		}()
 
		// non-blocking select on two channels see if the request
		// times out or finishes
		select {
 
		// if the context is done it timed out or was cancelled
		// so don't return anything
		case <-ctx.Done():
			return
 
			// if the request finished then finish the request by
			// writing the response
		case res := <-doneChan:
			c.JSON(res.status, res.body)
		}
	}
}

// https://gin-gonic.com/zh-cn/docs/examples/using-middleware/
func AuthRequired() func(c *gin.Context) {
	return func(c *gin.Context) {
		// get the underlying request context
		// ctx := c.Request.Context()
	}
}

func InitRouter() *gin.Engine {

	// 禁用控制台颜色，将日志写入文件时不需要控制台颜色。
	// gin.DisableConsoleColor()
	// 强制日志颜色化
	// gin.ForceConsoleColor()

	// 记录到文件。
	// f, _ := os.Create("log/tomato.log")
	// gin.DefaultWriter = io.MultiWriter(f)

	// 如果需要同时将日志写入文件和控制台，请使用以下代码。
	// gin.DefaultWriter = io.MultiWriter(f, os.Stdout)

	// 使用默认中间件（logger 和 recovery 中间件）创建 gin 路由
	router := gin.Default()
	// router := gin.New()

	// 定义路由日志的格式
	// https://gin-gonic.com/zh-cn/docs/examples/define-format-for-the-log-of-routes/
	// gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
	// 	log.Printf("endpoint %v %v %v %v\n", httpMethod, absolutePath, handlerName, nuHandlers)
	// }

	// 全局中间件
	// Logger 中间件将日志写入 gin.DefaultWriter，即使你将 GIN_MODE 设置为 release。
	// By default gin.DefaultWriter = os.Stdout
	router.Use(gin.Logger())

	// Recovery 中间件会 recover 任何 panic。如果有 panic 的话，会写入 500。
	router.Use(gin.Recovery())

	// add timeout middleware with 2 second duration
	router.Use(TimeoutMiddleware(time.Second * 2))

	//设置默认路由当访问一个错误网站时返回
	router.NoRoute(Controller.Error404)

	// 服务端要给客户端cookie
	router.GET("/cookie", func(c *gin.Context) {
		// 获取客户端是否携带cookie
		cookie,err := c.Cookie("gin_cookie")
		if err != nil {
			// 设置cookie
			c.SetCookie(
				"gin_cookie", // 设置cookie的key
				"test", // 设置cookie的值
				60, // 过期时间
				"/", // 所在目录
				"127.0.0.1",  //域名
				false,  // 是否只能通过https访问
				true) // 是否允许别人通过js获取自己的cookie
		}
		fmt.Println("cookie key:", cookie)
	})

	// store, _ := redis.NewStore(10, "tcp", "localhost:6379", "", []byte("secret"))
	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("mysession", store))
	store.Options(sessions.Options{
		MaxAge: int(1 * 1), //20*60 =30min
		Path:   "/",
	})

	// router.Static("/assets", "./assets")
	// router.StaticFS("/more_static", http.Dir("my_file_system"))
	// router.StaticFile("/favicon.ico", "./resources/favicon.ico")

	router.StaticFS("/webroot", http.Dir("remark")) // 静态文件路径
	// https://gin-gonic.com/zh-cn/docs/examples/html-rendering/
	router.LoadHTMLGlob("remark/web/**/*.html") //渲染html页面

	index := router.Group("/",  AuthRequired())
	{
		index.GET("/", Controller.IndexMain, TimedHandler(time.Second))
		index.POST("/",  Controller.IndexMain, TimedHandler(time.Second))
	}

	// router.GET("/", timedHandler(time.Second), Controller.AuthMiddleWare(), func(c *gin.Context){
	// 	c.JSON(200,gin.H{"message":"登录成功"})
	// })

	login := router.Group("/login")
	{
		User := new(Controller.User)
		login.GET("/", User.Login)
		login.POST("/", User.Login)
	}

	lock := router.Group("/lock")
	{
		User := new(Controller.User)
		lock.GET("/", User.LockSess)
		lock.POST("/", User.LockSess)
	}

	register := router.Group("/register")
	{
		User := new(Controller.User)
		register.GET("/", User.Register)
		register.POST("/", User.Register)
	}

	forgot := router.Group("/forgot")
	{
		forgot.GET("/", Controller.ForgotPass)
		forgot.POST("/", Controller.ForgotPass)
	}

	maintain := router.Group("/maintain")
	{
		maintain.GET("/", Controller.Maintain)
		maintain.POST("/", Controller.Maintain)
	}

	error400 := router.Group("/400")
	{
		error400.GET("/", Controller.Error400)
		error400.POST("/", Controller.Error400)
	}

	error403 := router.Group("/403")
	{
		error403.GET("/", Controller.Error403)
		error403.POST("/", Controller.Error403)
	}

	error404 := router.Group("/404")
	{
		error404.GET("/", Controller.Error404)
		error404.POST("/", Controller.Error404)
	}

	error500 := router.Group("/500")
	{
		error500.GET("/", Controller.Error500)
		error500.POST("/", Controller.Error500)
	}

	error503 := router.Group("/503")
	{
		error503.GET("/", Controller.Error503)
		error503.POST("/", Controller.Error503)
	}

	mobile := router.Group("/mobile")
	{
		Mobile := new(Controller.Mobile)
		mobile.GET("/", Mobile.Dial)
		mobile.POST("/", Mobile.Dial)
	}

	user := router.Group("/user")
	{
		User := new(Controller.User)
		user.GET("/login", User.Login)
		user.POST("/login", User.Login)
	}

	web := router.Group("/test")
	{
		adminUser := new(Controller.AdminUser)
		web.GET("/test", adminUser.Test)
		web.GET("/session_test", adminUser.SessionTest)
		web.GET("/template_test", adminUser.Template)
	}

	v1 := router.Group("/api/v1")
	{
		adminUser := new(Controller.AdminUser)
		v1.GET("/admin-users", adminUser.Index)
		v1.POST("/admin-users", adminUser.Store)
		v1.PATCH("/admin-users/:id", adminUser.Update)
		v1.DELETE("/admin-users/:id", adminUser.Destroy)
		v1.GET("/admin-users/:id", adminUser.Show)
	}
	return router
}
