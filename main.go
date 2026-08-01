package main

import (
	"fmt"
	"net/http"
	"news-restapi/handlers"
	"news-restapi/routes"
	"news-restapi/storage"
)

func main() {

	storage.ConnectDB()

	storage.InitTables()

	fmt.Println("Server is working on 8080 port")

	mux := routes.SetupRoutes()
	err := http.ListenAndServe(":8080", handlers.EnableCORS(mux))
	if err != nil {
		fmt.Println("Failed to start a server: ", err)
	}
}
