package routes

import (
	"net/http"
	"news-restapi/handlers"
)

func SetupRoutes() *http.ServeMux {

	mux := http.NewServeMux()

	setupNewsRoues(mux)
	setupUsersRoutes(mux)

	return mux
}

func setupNewsRoues(mux *http.ServeMux) {
	mux.HandleFunc("POST /news", handlers.AuthMiddleware(handlers.CreateNews))
	mux.HandleFunc("GET /news", handlers.GetNews)
	mux.HandleFunc("GET /news/{id}", handlers.GetNewsByID)
	mux.HandleFunc("PUT /news/{id}", handlers.AuthMiddleware(handlers.UpdateNews))
	mux.HandleFunc("PATCH /news/{id}", handlers.AuthMiddleware(handlers.PatchNews))
	mux.HandleFunc("DELETE /news/{id}", handlers.DeleteNews)
}

func setupUsersRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /auth/signup", handlers.CreateUser)
	mux.HandleFunc("POST /auth/login", handlers.LoginUser)
	mux.HandleFunc("GET /users", handlers.GetUsers)
	mux.HandleFunc("GET /users/{id}", handlers.GetUsersByID)
	mux.HandleFunc("PATCH /users/{id}", handlers.UpdateUser)
	mux.HandleFunc("DELETE /users/{id}", handlers.DeleteUser)
}
