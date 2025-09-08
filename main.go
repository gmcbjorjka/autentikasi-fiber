package main

import (
	"log"

	"autentikasi/database"
	"autentikasi/models"
	"autentikasi/routes"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	// connect DB
	database.ConnectDB()
	if err := database.DB.AutoMigrate(&models.User{}); err != nil {
		log.Fatal("Migration failed:", err)
	}

	// routes
	routes.AuthRoutes(app)

	// run server
	log.Fatal(app.Listen(":3000"))
}
