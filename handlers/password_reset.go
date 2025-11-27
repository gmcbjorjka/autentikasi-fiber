package handlers

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"

	"autentikasi/config"
	"autentikasi/database"
	"autentikasi/dto"
	"autentikasi/models"
	"autentikasi/utils"
)

// ForgotPassword - Step 1: Request OTP for password reset
// POST /api/v1/auth/forgot-password
func ForgotPassword(c *fiber.Ctx) error {
	cfg := c.Locals("config").(*config.Config)

	var req dto.ForgotPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "400",
			"message": "Invalid request body",
			"success": false,
		})
	}

	// Validate email format
	if req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "400",
			"message": "Email is required",
			"success": false,
		})
	}

	// Check if user exists
	var user models.User
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"code":    "404",
			"message": "Email not found",
			"success": false,
		})
	}

	// Generate OTP
	otp := utils.GenerateOTP()
	expiresAt := time.Now().Add(15 * time.Minute)

	// Save OTP to database
	passwordReset := models.PasswordReset{
		Email:     req.Email,
		OTP:       otp,
		ExpiresAt: expiresAt,
	}

	// Delete old OTP records for this email
	database.DB.Where("email = ?", req.Email).Delete(&models.PasswordReset{})

	if err := database.DB.Create(&passwordReset).Error; err != nil {
		log.Printf("[ForgotPassword] Failed to save OTP: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "500",
			"message": "Failed to generate OTP",
			"success": false,
		})
	}

	// Send OTP email
	if err := utils.SendOTPEmail(cfg, req.Email, otp); err != nil {
		log.Printf("[ForgotPassword] Failed to send email: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "500",
			"message": "Failed to send OTP email",
			"success": false,
		})
	}

	log.Printf("[ForgotPassword] OTP sent to %s", req.Email)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"code":    "200",
		"message": "OTP sent to your email",
		"success": true,
		"data": fiber.Map{
			"email":      req.Email,
			"expires_in": 15 * 60, // seconds
		},
	})
}

// VerifyOTP - Step 2: Verify OTP
// POST /api/v1/auth/verify-otp
func VerifyOTP(c *fiber.Ctx) error {
	var req dto.VerifyOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "400",
			"message": "Invalid request body",
			"success": false,
		})
	}

	// Validate inputs
	if req.Email == "" || req.OTP == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "400",
			"message": "Email and OTP are required",
			"success": false,
		})
	}

	// Check if user exists
	var user models.User
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"code":    "404",
			"message": "Email not found",
			"success": false,
		})
	}

	// Find OTP record
	var pwReset models.PasswordReset
	if err := database.DB.Where("email = ? AND otp = ?", req.Email, req.OTP).First(&pwReset).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    "401",
			"message": "Invalid OTP",
			"success": false,
		})
	}

	// Check if OTP is expired
	if time.Now().After(pwReset.ExpiresAt) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    "401",
			"message": "OTP has expired",
			"success": false,
		})
	}

	log.Printf("[VerifyOTP] OTP verified for %s", req.Email)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"code":    "200",
		"message": "OTP verified successfully",
		"success": true,
		"data": fiber.Map{
			"email": req.Email,
			"otp":   req.OTP,
		},
	})
}

// ResetPassword - Step 3: Reset password with valid OTP
// POST /api/v1/auth/reset-password
func ResetPassword(c *fiber.Ctx) error {
	cfg := c.Locals("config").(*config.Config)

	var req dto.ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "400",
			"message": "Invalid request body",
			"success": false,
		})
	}

	// Validate inputs
	if req.Email == "" || req.OTP == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "400",
			"message": "Email, OTP, and password are required",
			"success": false,
		})
	}

	if len(req.Password) < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "400",
			"message": "Password must be at least 6 characters",
			"success": false,
		})
	}

	// Check if user exists
	var user models.User
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"code":    "404",
			"message": "Email not found",
			"success": false,
		})
	}

	// Find and validate OTP record
	var pwReset models.PasswordReset
	if err := database.DB.Where("email = ? AND otp = ?", req.Email, req.OTP).First(&pwReset).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    "401",
			"message": "Invalid OTP",
			"success": false,
		})
	}

	// Check if OTP is expired
	if time.Now().After(pwReset.ExpiresAt) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    "401",
			"message": "OTP has expired",
			"success": false,
		})
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("[ResetPassword] Failed to hash password: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "500",
			"message": "Failed to reset password",
			"success": false,
		})
	}

	// Update user password
	if err := database.DB.Model(&user).Update("password", string(hashedPassword)).Error; err != nil {
		log.Printf("[ResetPassword] Failed to update password: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "500",
			"message": "Failed to reset password",
			"success": false,
		})
	}

	// Delete OTP record
	database.DB.Delete(&pwReset)

	// Send confirmation email
	if err := utils.SendPasswordResetSuccessEmail(cfg, req.Email); err != nil {
		log.Printf("[ResetPassword] Failed to send confirmation email: %v", err)
		// Don't fail the request if email fails, password is already updated
	}

	log.Printf("[ResetPassword] Password reset successfully for %s", req.Email)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"code":    "200",
		"message": "Password reset successfully",
		"success": true,
		"data": fiber.Map{
			"email": req.Email,
		},
	})
}
