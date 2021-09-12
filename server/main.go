package main

import (
	Config "tomato/config"
	Router "tomato/router"
	System "tomato/system"
)

func main() {
	cfg := Config.GetGlobalConfig()

	System.InitializeSystem(cfg)

	System.InitializeWebServer(cfg.WebServerAddr, Router.InitializeRouter(cfg))
}
