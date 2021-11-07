package system

import (
	"fmt"
	"log"
	"os"
	"runtime"
	Global "vm_manager/vm_admin/global"

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

func checkEnvironment() {
	// if occupied, sockets, err := ports.IsPortOccupied([]string{v2rayAListeningPort + ":tcp"}); occupied {
	// 	if err != nil {
	// 		log.Fatal("netstat:", err)
	// 	}
	// 	for _, socket := range sockets {
	// 		process, err := socket.Process()
	// 		if err == nil {
	// 			log.Fatal("Port %v is occupied by %v/%v", v2rayAListeningPort, process.Name, process.PID)
	// 		}
	// 	}
	// }

	// //等待网络连通
	// v2ray.CheckAndStopTransparentProxy()
	// for {
	// 	addrs, err := resolv.LookupHost("apple.com")
	// 	if err == nil && len(addrs) > 0 {
	// 		break
	// 	}
	// 	log.Alert("waiting for network connected")
	// 	time.Sleep(5 * time.Second)
	// }
	// log.Alert("network is connected")

}

func InitializeSystem(env *Global.Environment) {
	checkEnvironment()
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
