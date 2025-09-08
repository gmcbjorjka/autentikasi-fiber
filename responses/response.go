package responses

import "github.com/gofiber/fiber/v2"

func SuccessResponse(code, message string, data interface{}) fiber.Map {
	return fiber.Map{
		"code":    code,
		"message": message,
		"data":    data,
		"success": true,
	}
}

func ErrorResponse(code, message string) fiber.Map {
	return fiber.Map{
		"code":    code,
		"message": message,
		"data":    nil,
		"success": false,
	}
}
