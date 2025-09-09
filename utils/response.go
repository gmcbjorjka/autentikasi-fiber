package utils

import (
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func Ok(c *fiber.Ctx, status int, data interface{}) error {
	return c.Status(status).JSON(fiber.Map{
		"code":    "200",
		"message": "Success",
		"data":    data,
		"success": true,
	})
}

func Fail(c *fiber.Ctx, status int, msg string) error {
	return c.Status(status).JSON(fiber.Map{
		"code":    "500",
		"message": msg,
		"data":    nil,
		"success": false,
	})
}

func FiberErrorHandler(c *fiber.Ctx, err error) error {
	return Fail(c, fiber.StatusInternalServerError, err.Error())
}

func SignJWT(secret string, claims jwt.MapClaims) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(secret))
}
