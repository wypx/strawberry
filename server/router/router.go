package router

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"
	Config "tomato/config"

	"github.com/fatih/color"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"go.uber.org/ratelimit"
)

type StrawberryContext struct {
	Router *gin.Engine

	// This mutex protect Keys map
	KeysMutex *sync.RWMutex

	// Keys is a key/value pair exclusively for the context of each request.
	Keys map[string]interface{}
}

var StawCtx StrawberryContext

// Set is used to store a new key/value pair exclusively for this context.
// It also lazy initializes  c.Keys if it was not used previously.
func (c *StrawberryContext) Set(key string, value interface{}) {
	if c.KeysMutex == nil {
		c.KeysMutex = &sync.RWMutex{}
	}

	c.KeysMutex.Lock()
	if c.Keys == nil {
		c.Keys = make(map[string]interface{})
	}

	c.Keys[key] = value
	c.KeysMutex.Unlock()
}

// Get returns the value for the given key, ie: (value, true).
// If the value does not exists it returns (nil, false)
func (c *StrawberryContext) Get(key string) (value interface{}, exists bool) {
	if c.KeysMutex == nil {
		c.KeysMutex = &sync.RWMutex{}
	}

	c.KeysMutex.RLock()
	value, exists = c.Keys[key]
	c.KeysMutex.RUnlock()
	return
}

// MustGet returns the value for the given key if it exists, otherwise it panics.
func (c *StrawberryContext) MustGet(key string) interface{} {
	if value, exists := c.Get(key); exists {
		return value
	}
	panic("Key \"" + key + "\" does not exist")
}

// GetString returns the value associated with the key as a string.
func (c *StrawberryContext) GetString(key string) (s string) {
	if val, ok := c.Get(key); ok && val != nil {
		s, _ = val.(string)
	}
	return
}

// GetBool returns the value associated with the key as a boolean.
func (c *StrawberryContext) GetBool(key string) (b bool) {
	if val, ok := c.Get(key); ok && val != nil {
		b, _ = val.(bool)
	}
	return
}

// GetInt returns the value associated with the key as an integer.
func (c *StrawberryContext) GetInt(key string) (i int) {
	if val, ok := c.Get(key); ok && val != nil {
		i, _ = val.(int)
	}
	return
}

// GetInt64 returns the value associated with the key as an integer.
func (c *StrawberryContext) GetInt64(key string) (i64 int64) {
	if val, ok := c.Get(key); ok && val != nil {
		i64, _ = val.(int64)
	}
	return
}

// GetFloat64 returns the value associated with the key as a float64.
func (c *StrawberryContext) GetFloat64(key string) (f64 float64) {
	if val, ok := c.Get(key); ok && val != nil {
		f64, _ = val.(float64)
	}
	return
}

func SetStrawberryRouter(router *gin.Engine) {
	StawCtx.Router = router
}

func RedirectLogin(c *gin.Context) {
	c.Request.URL.Path = "/login"
	/* Router redirect, using HandleContext */
	StawCtx.Router.HandleContext(c)

	// c.Redirect(http.StatusMovedPermanently, "luotang.me")
	// c.Redirect(http.StatusFound, "/foo")
}

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
		type ResponseData struct {
			status int
			body   map[string]interface{}
		}

		// create a done channel to tell the request it's done
		doneChan := make(chan ResponseData)

		// here you put the actual work needed for the request
		// and then send the doneChan with the status and body
		// to finish the request by writing the response
		go func() {
			time.Sleep(duration)
			doneChan <- ResponseData{
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
	log.Printf(color.CyanString("Current AuthRequired"))
	return func(c *gin.Context) {
		// get the underlying request context
		// ctx := c.Request.Context()
	}
}

type myForm struct {
	Colors []string `form:"colors[]"`
}

func indexHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "form.html", nil)
}

func formHandler(c *gin.Context) {
	var fakeForm myForm
	c.Bind(&fakeForm)
	c.JSON(http.StatusOK, gin.H{"color": fakeForm.Colors})
}

var (
	limit ratelimit.Limiter
	rps   = 100
)

func LeakBucket() gin.HandlerFunc {
	prev := time.Now()
	return func(ctx *gin.Context) {
		now := limit.Take()
		log.Print(color.CyanString("%v", now.Sub(prev)))
		prev = now
	}
}

func InitializeRouter(cfg *Config.GlobalConfig) *gin.Engine {
	// https://gin-gonic.com/zh-cn/docs/examples/run-multiple-service/
	router := gin.Default()

	SetStrawberryRouter(router)

	// 定义路由日志的格式
	// https://gin-gonic.com/zh-cn/docs/examples/define-format-for-the-log-of-routes/
	gin.DebugPrintRouteFunc = func(mothod, abspath, handlename string, nuHandlers int) {
		log.Printf("endpoint %v %v %v %v\n", mothod, abspath, handlename, nuHandlers)
	}

	// router.Use(LeakBucket())

	router.GET("/rate", func(ctx *gin.Context) {
		ctx.JSON(200, "rate limiting test")
	})

	log.Printf(color.CyanString("Current Rate Limit: %v requests/s", rps))

	// 全局中间件
	// Logger 中间件将日志写入 gin.DefaultWriter，即使你将 GIN_MODE 设置为 release。
	// By default gin.DefaultWriter = os.Stdout
	router.Use(gin.Logger())

	// Recovery 中间件会 recover 任何 panic。如果有 panic 的话，会写入 500。
	router.Use(gin.Recovery())

	// add timeout middleware with 2 second duration
	router.Use(TimeoutMiddleware(time.Second * 10))

	// router.Use(middleware.Cors()) // 如需跨域可以打开

	// https://github.com/gin-gonic/examples/blob/master/favicon/main.go
	// router.Use(favicon.New("./favicon.ico"))

	//设置默认路由当访问一个错误网站时返回
	router.NoRoute(Error404)
	router.NoMethod(Error404)

	// 服务端要给客户端cookie
	router.GET("/cookie", func(c *gin.Context) {
		// 获取客户端是否携带cookie
		cookie, err := c.Cookie("gin_cookie")
		if err != nil {
			// 设置cookie
			c.SetCookie(
				"gin_cookie", // 设置cookie的key
				"test",       // 设置cookie的值
				3600,         // 过期时间
				"/",          // 所在目录
				"127.0.0.1",  //域名
				false,        // 是否只能通过https访问
				true)         // 是否允许别人通过js获取自己的cookie
		}
		log.Println("cookie key:", cookie)
	})

	// store, _ := redis.NewStore(10, "tcp", "localhost:6379", "", []byte("secret"))
	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("mysession", store))
	store.Options(sessions.Options{
		MaxAge: int(1 * 1), // 20*60 =30min
		Path:   "/",
	})

	// router.Static("/assets", "./assets")
	// router.StaticFS("/more_static", http.Dir("my_file_system"))
	// router.StaticFile("/favicon.ico", "./resources/favicon.ico")

	router.StaticFS(cfg.WebRelativePath, http.Dir(cfg.WebAbsolutePath))
	// https://gin-gonic.com/zh-cn/docs/examples/html-rendering/
	// https://gin-gonic.com/zh-cn/docs/examples/multiple-template/
	router.LoadHTMLGlob("/root/work/strawberry/web/static/*.html") //渲染html页面

	// 认证路由组
	authorized := router.Group("/", AuthRequired())
	{
		authorized.GET("/", IndexMain, TimedHandler(time.Second))
		authorized.POST("/", IndexMain, TimedHandler(time.Second))
	}
	// 	login := router.Group("/login", AuthMiddleWare())
	login := router.Group("/login")
	{
		User := new(User)
		login.GET("/", User.Login)
		login.POST("/", User.Login)
	}

	lock := router.Group("/lock")
	{
		User := new(User)
		lock.GET("/", User.LockSess)
		lock.POST("/", User.LockSess)
	}

	register := router.Group("/register")
	{
		User := new(User)
		register.GET("/", User.Register)
		register.POST("/", User.Register)
	}

	forgot := router.Group("/forgot")
	{
		forgot.GET("/", ForgotPass)
		forgot.POST("/", ForgotPass)
	}

	maintain := router.Group("/maintain")
	{
		maintain.GET("/", Maintain)
		maintain.POST("/", Maintain)
	}

	error400 := router.Group("/400")
	{
		error400.GET("/", Error400)
		error400.POST("/", Error400)
	}

	error403 := router.Group("/403")
	{
		error403.GET("/", Error403)
		error403.POST("/", Error403)
	}

	error404 := router.Group("/404")
	{
		error404.GET("/", Error404)
		error404.POST("/", Error404)
	}

	error500 := router.Group("/500")
	{
		error500.GET("/", Error500)
		error500.POST("/", Error500)
	}

	error503 := router.Group("/503")
	{
		error503.GET("/", Error503)
		error503.POST("/", Error503)
	}

	mobile := router.Group("/mobile")
	{
		Mobile := new(Mobile)
		mobile.GET("/", Mobile.Dial)
		mobile.POST("/", Mobile.Dial)
	}

	user := router.Group("/user")
	{
		User := new(User)
		user.GET("/login", User.Login)
		user.POST("/login", User.Login)
	}

	router.GET("/test_form", indexHandler)
	router.POST("/test_form", formHandler)

	test := router.Group("/test")
	{
		adminUser := new(AdminUser)
		test.GET("/test", adminUser.Test)
		test.GET("/session_test", adminUser.SessionTest)
		test.GET("/template_test", adminUser.Template)
	}

	/* API Service */
	v1 := router.Group("/api/v1")
	{
		adminUser := new(AdminUser)
		v1.GET("/admin-users", adminUser.Index)
		v1.POST("/admin-users", adminUser.Store)
		v1.PATCH("/admin-users/:id", adminUser.Update)
		v1.DELETE("/admin-users/:id", adminUser.Destroy)
		v1.GET("/admin-users/:id", adminUser.Show)
		// v1.HEAD("/someHead", head)
		// v1.OPTIONS("/someOptions", options)
	}

	// Set a lower memory limit for multipart forms (default is 32 MiB)
	router.MaxMultipartMemory = 8 << 20 // 8 MiB
	up := router.Group("/upload")
	{
		upload := new(Upload)
		up.GET("/upload_single", upload.UploadSingleFile)
		up.GET("/upload_muti", upload.UploadMutiFile)
	}

	return router
}
