package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

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

// UpdateMe -> update current authenticated user's profile
func UpdateMe(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil {
		return utils.Fail(c, fiber.StatusUnauthorized, "Unauthorized")
	}

	var body struct {
		Nama     string `json:"nama"`
		ImgURL   string `json:"img"`
		Phone    string `json:"phone"`
		Gender   string `json:"gender"`
		Birthday string `json:"birthday"` // ISO date
		Status   string `json:"status"`
	}
	if err := c.BodyParser(&body); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "Invalid JSON body")
	}

	// apply updates if provided
	if body.Nama != "" {
		user.Nama = body.Nama
	}
	if body.ImgURL != "" {
		user.ImgURL = &body.ImgURL
	}
	if body.Phone != "" {
		user.Phone = &body.Phone
	}
	if body.Gender != "" {
		user.Gender = &body.Gender
	}
	if body.Status != "" {
		user.Status = &body.Status
	}
	if body.Birthday != "" {
		if t, err := time.Parse(time.RFC3339, body.Birthday); err == nil {
			user.Birthday = &t
		}
	}

	if err := database.DB.Save(user).Error; err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "Failed to update profile")
	}

	// reuse Me response to include balance
	return Me(c)
}

// UploadAvatar -> accept multipart file field `avatar`, save it to ./uploads
// and update user's ImgURL to the served path.
func UploadAvatar(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil {
		return utils.Fail(c, fiber.StatusUnauthorized, "Unauthorized")
	}

	file, err := c.FormFile("avatar")
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "avatar file is required")
	}

	// ensure uploads dir
	if err := os.MkdirAll("uploads", 0755); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "Failed to prepare upload directory")
	}

	ext := filepath.Ext(file.Filename)
	fname := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	dest := filepath.Join("uploads", fname)
	if err := c.SaveFile(file, dest); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "Failed to save uploaded file")
	}

	// Build a URL that the client can use. Use BaseURL() so it contains scheme+host.
	imgURL := c.BaseURL() + "/uploads/" + fname
	user.ImgURL = &imgURL
	if err := database.DB.Save(user).Error; err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "Failed to save user avatar")
	}

	return utils.Ok(c, fiber.StatusOK, fiber.Map{"img": imgURL})
}
