package main

import (
	"fmt"
	// "log"
	"tomato/config"
	// "tomato/database"
	"tomato/routes"
)

func main() {
	if err := config.Load("config/config.yaml"); err != nil {
		fmt.Println("Failed to load configuration")
		return
	}

	// db, err := database.InitDB()
	// if err != nil {
	// 	fmt.Println("err open databases")
	// 	return
	// }
	// defer db.Close()

	router := routes.InitRouter()
	router.Run(config.Get().Addr)

	// https://gin-gonic.com/zh-cn/docs/examples/graceful-restart-or-stop/
	// 等待中断信号以优雅地关闭服务器（设置 5 秒的超时时间）
	// quit := make(chan os.Signal)
	// signal.Notify(quit, os.Interrupt)
	// <-quit
	// log.Println("Shutdown Server ...")

	// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cancel()
	// if err := srv.Shutdown(ctx); err != nil {
	// 	log.Fatal("Server Shutdown:", err)
	// }
	// log.Println("Server exiting")
}
