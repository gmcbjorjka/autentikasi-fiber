package middleware

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"autentikasi/config"
	"autentikasi/database"
	"autentikasi/models"
)

func JWTProtected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"code": "401", "message": "Missing token", "data": nil, "success": false})
		}
		tokenStr := strings.TrimPrefix(auth, "Bearer ")

		cfg := config.Load()
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(cfg.JWTSecret), nil
		})
		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"code": "401", "message": "Invalid token", "data": nil, "success": false})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"code": "401", "message": "Invalid claims", "data": nil, "success": false})
		}

		var user models.User
		if err := database.DB.Where("id = ?", claims["sub"]).First(&user).Error; err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"code": "401", "message": "User not found", "data": nil, "success": false})
		}
		c.Locals("user", &user)
		return c.Next()
	}
}
