package routes

import (
	"autentikasi/controllers"

	"github.com/gofiber/fiber/v2"
)

func AuthRoutes(app *fiber.App) {
	api := app.Group("/api/v1")

	api.Post("/register", controllers.Register)
	api.Post("/login", controllers.Login)
}
