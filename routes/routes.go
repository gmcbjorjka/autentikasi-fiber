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
	api.Put("/me", middleware.JWTProtected(), handlers.UpdateMe)
	api.Post("/me/avatar", middleware.JWTProtected(), handlers.UploadAvatar)

	// Transactions
	api.Get("/transactions", middleware.JWTProtected(), handlers.ListTransactions)
	api.Post("/transactions", middleware.JWTProtected(), handlers.CreateTransaction)

	// Friends
	friends := api.Group("/friends", middleware.JWTProtected())
	friends.Get("/search", handlers.SearchUserByPhone)
	friends.Post("/request", handlers.SendFriendRequest)
	friends.Post("/accept/:id", handlers.AcceptFriendRequest)
	friends.Post("/reject/:id", handlers.RejectFriendRequest)
	friends.Get("/list", handlers.ListFriends)
	friends.Get("/pending", handlers.ListPendingRequests)
	friends.Delete("/:id", handlers.DeleteFriend)
	friends.Post("/:id/toggle-debt", handlers.ToggleDebt)

}
