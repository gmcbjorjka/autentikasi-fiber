package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"

	"autentikasi/config"
	"autentikasi/database"
	"autentikasi/routes"
	"autentikasi/utils"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	if err := database.Connect(cfg); err != nil {
		log.Fatalf("DB connect error: %v", err)
	}
	log.Println("✅ Database connected")

	app := fiber.New(fiber.Config{ErrorHandler: utils.FiberErrorHandler})
	app.Use(logger.New())
	app.Use(cors.New())

	// Register routes
	routes.SetupRoutes(app)

	addr := ":" + cfg.AppPort
	log.Printf("🚀 Auth Service running on http://localhost%v\n", addr)
	if err := app.Listen(addr); err != nil {
		log.Println("server stopped:", err)
	}
}
