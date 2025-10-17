package handlers

import (
	"autentikasi/database"
	"autentikasi/models"
	"autentikasi/utils"

	"github.com/gofiber/fiber/v2"
)

// CreateTransaction creates a new transaction for the authorized user
func CreateTransaction(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil {
		return utils.Fail(c, fiber.StatusUnauthorized, "Unauthorized")
	}

	var body struct {
		Jenis      string  `json:"jenis"`
		Kategori   string  `json:"kategori"`
		Jumlah     float64 `json:"jumlah"`
		Metode     string  `json:"metode"`
		Keterangan string  `json:"keterangan"`
	}
	if err := c.BodyParser(&body); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "Invalid body")
	}

	tx := models.Transaction{
		UserID:     user.ID,
		Jenis:      body.Jenis,
		Kategori:   body.Kategori,
		Jumlah:     body.Jumlah,
		Metode:     body.Metode,
		Keterangan: body.Keterangan,
	}
	if err := database.DB.Create(&tx).Error; err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "Failed to create transaction")
	}

	return utils.Ok(c, fiber.StatusCreated, tx)
}

// ListTransactions returns transactions of the authenticated user
func ListTransactions(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil {
		return utils.Fail(c, fiber.StatusUnauthorized, "Unauthorized")
	}

	var txs []models.Transaction
	if err := database.DB.Where("user_id = ?", user.ID).Order("created_at desc").Find(&txs).Error; err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "Failed to fetch transactions")
	}

	return utils.Ok(c, fiber.StatusOK, txs)
}
