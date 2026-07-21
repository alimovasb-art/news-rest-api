package main

import (
	"fmt"
	"net/http"
	"news-restapi/routes"
)

func main() {
	fmt.Println("Server is working on 8080 port")

	mux := routes.SetupRoutes()
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		fmt.Println("Failed to start a server: ", err)
	}
}
