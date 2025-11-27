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

	// Optional dev seeding: set DEV_SEED=true in your environment to create
	// a developer user (dev@example.com / password) and sample transactions.
	if err := database.SeedDev(); err != nil {
		log.Printf("[SEED] error: %v", err)
	}

	app := fiber.New(fiber.Config{ErrorHandler: utils.FiberErrorHandler})
	app.Use(logger.New())
	app.Use(cors.New())

	// serve uploaded files under /uploads
	app.Static("/uploads", "./uploads")

	// Register routes (pass config so handlers can access via c.Locals("config"))
	routes.SetupRoutes(app, cfg)

	addr := cfg.AppHost + ":" + cfg.AppPort
	log.Printf("🚀 Auth Service running on http://%s:%s\n", cfg.AppHost, cfg.AppPort)
	if err := app.Listen(addr); err != nil {
		log.Println("server stopped:", err)
	}
}
