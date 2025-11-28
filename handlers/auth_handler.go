package handlers

import (
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"autentikasi/config"
	"autentikasi/database"
	"autentikasi/dto"
	"autentikasi/models"
	"autentikasi/utils"
)

func Register(c *fiber.Ctx) error {
	var body dto.RegisterRequest
	if err := c.BodyParser(&body); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "Invalid JSON body")
	}
	body.Email = strings.TrimSpace(strings.ToLower(body.Email))

	// Basic validation
	if body.Nama == "" {
		return utils.Fail(c, fiber.StatusBadRequest, "Nama pengguna wajib diisi")
	}
	if body.Email == "" {
		return utils.Fail(c, fiber.StatusBadRequest, "E-mail wajib diisi")
	}
	if !strings.Contains(body.Email, "@") || !strings.Contains(body.Email, ".") {
		return utils.Fail(c, fiber.StatusBadRequest, "Format e-mail tidak valid")
	}
	if len(body.Password) < 6 {
		return utils.Fail(c, fiber.StatusBadRequest, "Kata sandi minimal 6 karakter")
	}

	var exists int64
	database.DB.Model(&models.User{}).Where("email = ?", body.Email).Count(&exists)
	if exists > 0 {
		return utils.Fail(c, fiber.StatusConflict, "E-mail sudah terdaftar")
	}

	hashPass, _ := utils.Hash(body.Password)

	user := models.User{
		Nama:     body.Nama,
		Email:    body.Email,
		Password: hashPass,
	}

	// PIN opsional
	if body.Pin != "" {
		hashPin, _ := utils.Hash(body.Pin)
		user.Pin = &hashPin
	}

	// ImgURL opsional
	if body.ImgURL != "" {
		user.ImgURL = &body.ImgURL
	}

	if err := database.DB.Create(&user).Error; err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "Gagal membuat pengguna")
	}

	return utils.Ok(c, fiber.StatusCreated, fiber.Map{
		"id":    user.ID,
		"nama":  user.Nama,
		"email": user.Email,
		"img":   user.ImgURL,
	})
}

func Login(c *fiber.Ctx) error {
	var body dto.LoginRequest
	if err := c.BodyParser(&body); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "Invalid JSON body")
	}
	email := strings.TrimSpace(strings.ToLower(body.Email))

	// Basic validation
	if email == "" {
		return utils.Fail(c, fiber.StatusBadRequest, "E-mail wajib diisi")
	}
	if body.Password == "" {
		return utils.Fail(c, fiber.StatusBadRequest, "Kata sandi wajib diisi")
	}

	var user models.User
	if err := database.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "E-mail kamu tidak ditemukan")
	}
	if ok := utils.Check(body.Password, user.Password); !ok {
		return utils.Fail(c, fiber.StatusUnauthorized, "Kata sandi kamu salah")
	}

	cfg := config.Load()
	claims := jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email,
		"nama":  user.Nama,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	}
	token, err := utils.SignJWT(cfg.JWTSecret, claims)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "Failed to sign token")
	}

	// Record login event
	go func(u models.User) {
		now := time.Now()
		ah := models.AccountHistory{
			UserID:      u.ID,
			Event:       "login",
			Description: "User logged in",
			CreatedAt:   &now,
		}
		if err := database.DB.Create(&ah).Error; err != nil {
			// log but don't fail
			log.Printf("[AccountHistory] failed to record login: %v", err)
		}
	}(user)

	return utils.Ok(c, fiber.StatusOK, fiber.Map{
		"token": token,
	})
}

func Me(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil {
		return utils.Fail(c, fiber.StatusUnauthorized, "Unauthorized")
	}
	// calculate balance from transactions
	var income float64
	var expense float64
	// sum pemasukan
	database.DB.Raw(
		"SELECT COALESCE(SUM(jumlah),0) FROM transactions WHERE user_id = ? AND jenis = ?",
		user.ID, "pemasukan",
	).Scan(&income)
	// sum pengeluaran
	database.DB.Raw(
		"SELECT COALESCE(SUM(jumlah),0) FROM transactions WHERE user_id = ? AND jenis = ?",
		user.ID, "pengeluaran",
	).Scan(&expense)
	balance := int64(income - expense)

	// optional fields
	var birthdayStr *string
	if user.Birthday != nil {
		s := user.Birthday.Format(time.RFC3339)
		birthdayStr = &s
	}

	return utils.Ok(c, fiber.StatusOK, fiber.Map{
		"id":       user.ID,
		"nama":     user.Nama,
		"email":    user.Email,
		"img":      user.ImgURL,
		"balance":  balance,
		"phone":    user.Phone,
		"gender":   user.Gender,
		"birthday": birthdayStr,
		"status":   user.Status,
	})
}

// Logout - record logout event (token should be discarded client-side)
func Logout(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil {
		return utils.Fail(c, fiber.StatusUnauthorized, "Unauthorized")
	}
	now := time.Now()
	ah := models.AccountHistory{
		UserID:      user.ID,
		Event:       "logout",
		Description: "User logged out",
		CreatedAt:   &now,
	}
	if err := database.DB.Create(&ah).Error; err != nil {
		log.Printf("[AccountHistory] failed to record logout: %v", err)
	}
	return utils.Ok(c, fiber.StatusOK, fiber.Map{"message": "Logged out"})
}

// GetAuthHistory - return recent account events for current user
func GetAuthHistory(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil {
		return utils.Fail(c, fiber.StatusUnauthorized, "Unauthorized")
	}
	var items []models.AccountHistory
	if err := database.DB.Where("user_id = ?", user.ID).Order("created_at desc").Find(&items).Error; err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "Failed to fetch history")
	}
	return utils.Ok(c, fiber.StatusOK, fiber.Map{"events": items})
}
