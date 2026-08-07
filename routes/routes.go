package routes

import (
	"news-restapi/handlers"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	setupNewsRoutes(app)
	setupUsersRoutes(app)
	setupStaticRoutes(app)
}

func setupNewsRoutes(app *fiber.App) {
	app.Post("/news", handlers.AuthMiddleware, handlers.CreateNews)
	app.Get("/news", handlers.GetNews)
	app.Get("/news/:id", handlers.GetNewsByID)
	app.Patch("/news/:id", handlers.AuthMiddleware, handlers.PatchNews)
	app.Delete("/news/:id", handlers.AuthMiddleware, handlers.DeleteNews)
}

func setupUsersRoutes(app *fiber.App) {
	app.Post("/auth/signup", handlers.CreateUser)
	app.Post("/auth/login", handlers.LoginUser)
	app.Get("/auth/me", handlers.AuthMiddleware, handlers.GetMe)
	app.Get("/users", handlers.GetUsers)
	app.Get("/users/:id", handlers.GetUsersByID)
	app.Patch("/users", handlers.AuthMiddleware, handlers.UpdateUser)
	app.Delete("/users/:id", handlers.AuthMiddleware, handlers.DeleteUser)
}

func setupStaticRoutes(app *fiber.App) {
	app.Static("/uploads", "./uploads")
}
