package routes

import (
	"github.com/gofiber/fiber/v2"

	"autentikasi/handlers"
	"autentikasi/middleware"
)

func SetupRoutes(app *fiber.App) {
	api := app.Group("/api/v1")

	auth := api.Group("/auth")
	auth.Post("/register", handlers.Register)
	auth.Post("/login", handlers.Login)

	api.Get("/me", middleware.JWTProtected(), handlers.Me)
	api.Get("/users/:id", middleware.JWTProtected(), handlers.GetUserProfile)

	// Transactions
	api.Get("/transactions", middleware.JWTProtected(), handlers.ListTransactions)
	api.Post("/transactions", middleware.JWTProtected(), handlers.CreateTransaction)

}
