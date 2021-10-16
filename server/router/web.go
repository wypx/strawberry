package router

// func InitWebRouter(develop bool) *gin.Engine {
// 	var router *gin.Engine
// 	// 非调试模式（生产模式） 日志写到日志文件
// 	if develop {
// 		// 调试模式，开启 pprof 包，便于开发阶段分析程序性能
// 		router = gin.Default()
// 		pprof.Register(router)
// 	} else {
// 		//1.将日志写入日志文件
// 		gin.DisableConsoleColor()
// 		f, _ := os.Create("strawberry.log")
// 		gin.DefaultWriter = io.MultiWriter(f)
// 		// 2.如果是有nginx前置做代理，基本不需要gin框架记录访问日志，
// 		// 开启下面一行代码，屏蔽上面的三行代码，性能提升 5%
// 		//gin.SetMode(gin.ReleaseMode)
// 		router = gin.Default()
// 	}

// 	// 根据配置进行设置跨域
// 	if variable.ConfigYml.GetBool("HttpServer.AllowCrossDomain") {
// 		router.Use(cors.Next())
// 	}

// 	//处理静态资源（不建议gin框架处理静态资源，参见 public/readme.md 说明 ）
// 	router.Static("/public", "./public")             //  定义静态资源路由与实际目录映射关系
// 	router.StaticFS("/dir", http.Dir("./public"))    // 将public目录内的文件列举展示
// 	router.StaticFile("/abcd", "./public/readme.md") // 可以根据文件名绑定需要返回的文件名

// }
