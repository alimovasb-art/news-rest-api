package main

import (
	"fmt"
	"net/http"
	"news-restapi/handlers"
	"news-restapi/routes"
	"news-restapi/storage"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	storage.ConnectDB()

	storage.InitTables()

	fmt.Println("Server is working on 8080 port")

	mux := routes.SetupRoutes()
	err := http.ListenAndServe(":8080", handlers.EnableCORS(mux))
	if err != nil {
		fmt.Println("Failed to start a server: ", err)
	}
}
