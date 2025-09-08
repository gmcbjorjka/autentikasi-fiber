package controllers

import (
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"autentikasi/database"
	"autentikasi/models"
	"autentikasi/responses"
)

func Register(c *fiber.Ctx) error {
	var input models.User
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse("400", "Invalid request"))
	}

	// hash password
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(input.Password), 12)
	input.Password = string(hashedPassword)

	// hash pin
	hashedPin, _ := bcrypt.GenerateFromPassword([]byte(input.Pin), 12)
	input.Pin = string(hashedPin)

	if err := database.DB.Create(&input).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse("500", "Failed to register user"))
	}

	return c.Status(fiber.StatusOK).JSON(responses.SuccessResponse("200", "Register Success", nil))
}

func Login(c *fiber.Ctx) error {
	var input struct {
		Email string `json:"email"`
		Pin   string `json:"pin"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse("400", "Invalid request"))
	}

	var user models.User
	if err := database.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse("500", "Failed, User not registered"))
	}

	// check pin
	if err := bcrypt.CompareHashAndPassword([]byte(user.Pin), []byte(input.Pin)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(responses.ErrorResponse("401", "Invalid credentials"))
	}

	secret := os.Getenv("JWT_SECRET")
	expireTime, _ := time.ParseDuration(os.Getenv("JWT_EXPIRE") + "s")

	claims := jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email,
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(expireTime).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))

	data := fiber.Map{
		"accessToken":           tokenString,
		"accessTokenType":       "Bearer",
		"accessTokenExpiresSec": int(expireTime.Seconds()),
	}

	return c.Status(fiber.StatusOK).JSON(responses.SuccessResponse("200", "Success", data))
}
