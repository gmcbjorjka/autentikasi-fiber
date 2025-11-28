package routes

import (
	"github.com/gofiber/fiber/v2"

	"autentikasi/config"
	"autentikasi/handlers"
	"autentikasi/middleware"
)

func SetupRoutes(app *fiber.App, cfg *config.Config) {
	// Static files - serve uploads directory
	app.Static("/uploads", "./uploads")

	// Attach config to each request context so handlers can access it via c.Locals("config").
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("config", cfg)
		return c.Next()
	})

	api := app.Group("/api/v1")

	auth := api.Group("/auth")
	auth.Post("/register", handlers.Register)
	auth.Post("/login", handlers.Login)
	auth.Post("/logout", middleware.JWTProtected(), handlers.Logout)
	auth.Post("/forgot-password", handlers.ForgotPassword)
	auth.Post("/verify-otp", handlers.VerifyOTP)
	auth.Post("/reset-password", handlers.ResetPassword)
	auth.Get("/history", middleware.JWTProtected(), handlers.GetAuthHistory)

	api.Get("/me", middleware.JWTProtected(), handlers.Me)
	api.Get("/users/:id", middleware.JWTProtected(), handlers.GetUserProfile)
	api.Put("/me", middleware.JWTProtected(), handlers.UpdateMe)
	api.Post("/me/avatar", middleware.JWTProtected(), handlers.UploadAvatar)

	// Transactions
	api.Get("/transactions", middleware.JWTProtected(), handlers.ListTransactions)
	api.Post("/transactions", middleware.JWTProtected(), handlers.CreateTransaction)

	// Friends
	friends := api.Group("/friends", middleware.JWTProtected())
	friends.Get("/search", handlers.SearchUserByPhone)
	friends.Post("/request", handlers.SendFriendRequest)
	friends.Post("/accept", handlers.AcceptFriendRequestByPhone) // Accept with phone in body
	friends.Post("/reject", handlers.RejectFriendRequestByPhone) // Reject with phone in body
	friends.Post("/accept/:id", handlers.AcceptFriendRequest)    // Legacy: by ID
	friends.Post("/reject/:id", handlers.RejectFriendRequest)    // Legacy: by ID
	friends.Get("/list", handlers.ListFriends)
	friends.Get("/pending", handlers.ListPendingRequests)
	friends.Delete("/:id", handlers.DeleteFriend)
	friends.Post("/:id/toggle-debt", handlers.ToggleDebt)

}
