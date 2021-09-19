package system

import (
	"fmt"
	"log"
	"os"
	"runtime"
	Global "tomato/global"

	"github.com/gin-gonic/gin"
)

func InitializeEnv() {
	if runtime.GOOS == "windows" {
		log.Printf("Strawberry cannot run on windows")
		log.Printf("Press any key to continue...")
		_, _ = fmt.Scanf("\n")
		os.Exit(1)
	}
}

func InitializeLog() {
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

func InitializeSystem(env *Global.Environment) {
	InitializeEnv()
	InitializeLog()

	if env.WebServerEnv() == "develop" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	nuCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(nuCPU)
	log.Printf("Running with %d CPUs\n", nuCPU)
}
