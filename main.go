package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
	"tomato/routes"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
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

func InitializeWebServer(addr string, router *gin.Engine) {
	// ref: https://gin-gonic.com/zh-cn/docs/examples/graceful-restart-or-stop/
	// https://github.com/gin-gonic/examples/tree/master/graceful-shutdown
	// ref: https://www.cnblogs.com/cjyangblog/p/14695850.html
	srv := &http.Server{
		Addr:           ":8080",
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		log.Printf("Listening and serving HTTP on %s\n", srv.Addr)
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
		// err = srv.ListenAndServeTLS("./ssl/server.pem", "./ssl/server.key")
		// if err != nil && err != http.ErrServerClosed {
		// 	log.Fatalf("listen: %s\n", err)
		// }
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	log.Println("Server exiting")
}

type Config struct {
	Addr        string `yaml:"addr"`
	DSN         string `yaml:"dsn"`
	MaxIdleConn int    `yaml:"max_idle_conn"`
}

func GetConfig() *Config {
	result, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Printf("Failed to read configuration")
		return nil
	}
	var cfg *Config
	err = yaml.Unmarshal(result, &cfg)
	if err != nil {
		log.Printf("Failed to load configuration")
		return nil
	}

	if cfg.Addr == "" {
		port := os.Getenv("PORT")
		if port == "" {
			cfg.Addr = ":8080"
			log.Printf("Defaulting to port %s", port)
		} else {
			cfg.Addr = ":" + port
		}
	}

	return cfg
}

// ConfigRuntime sets the number of operating system threads.
func ConfigRuntime() {
	nuCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(nuCPU)
	fmt.Printf("Running with %d CPUs\n", nuCPU)
}

func main() {
	ConfigRuntime()

	InitializeLogger()

	// gin.SetMode(gin.ReleaseMode)

	cfg := GetConfig()

	InitializeWebServer(cfg.Addr, routes.InitializeRouter())
}
