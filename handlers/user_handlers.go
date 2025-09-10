package handlers

import (
	"strconv"

	"autentikasi/database"
	"autentikasi/models"
	"autentikasi/utils"

	"github.com/gofiber/fiber/v2"
)

// GetUserProfile -> ambil profile user by ID
func GetUserProfile(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "Invalid user ID")
	}

	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "User not found")
	}

	return utils.Ok(c, fiber.StatusOK, fiber.Map{
		"id":     user.ID,
		"nama":   user.Nama,
		"email":  user.Email,
		"img":    user.ImgURL,
		"joined": user.CreatedAt,
	})
}
