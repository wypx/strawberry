package main

import (
	Global "tomato/global"
	Router "tomato/router"
	System "tomato/system"
)

func main() {
	env := Global.GetEnvironmentConfig()

	System.InitializeSystem(env)

	System.InitializeWebServer(env.WebServerAddr(), Router.InitializeRouter(env))
}
