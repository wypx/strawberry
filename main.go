package main

import (
	"fmt"
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
}
