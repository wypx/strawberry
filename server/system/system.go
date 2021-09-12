package system

import (
	"log"
	"runtime"
	Config "tomato/config"

	"github.com/gin-gonic/gin"
)

func InitializeLogger() {
	// 禁用控制台颜色，将日志写入文件时不需要控制台颜色。
	// gin.DisableConsoleColor()
	// 强制日志颜色化
	// gin.ForceConsoleColor()

	// 记录到文件。
	// f, _ := os.Create("log/tomato.log")
	// gin.DefaultWriter = io.MultiWriter(f)

	// 如果需要同时将日志写入文件和控制台，请使用以下代码。
	// gin.DefaultWriter = io.MultiWriter(f, os.Stdout)

	// log.SetFlags(0)
	// log.SetPrefix("[GIN] ")
	// log.SetOutput(gin.DefaultWriter)
}

func InitializeSystem(cfg *Config.GlobalConfig) {
	InitializeLogger()

	if cfg.WebServerEnv == "develop" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	nuCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(nuCPU)
	log.Printf("Running with %d CPUs\n", nuCPU)
}
