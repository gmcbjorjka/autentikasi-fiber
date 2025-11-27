package handlers

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"

	"autentikasi/config"
	"autentikasi/utils"
)

// TestSMTPConnection - Test SMTP connection without sending actual email
// GET /api/v1/auth/test-smtp
func TestSMTPConnection(c *fiber.Ctx) error {
	cfg := c.Locals("config").(*config.Config)

	log.Println("[TestSMTP] Testing SMTP connection...")
	log.Printf("[TestSMTP] Config - Server: %s, Port: %s, Username: %s, Sender: %s",
		cfg.MailServer, cfg.MailPort, cfg.MailUsername, cfg.MailDefaultSender)

	// Test email address
	testEmail := "test@example.com"
	testOTP := "123456"

	err := utils.SendOTPEmail(cfg, testEmail, testOTP)
	if err != nil {
		log.Printf("[TestSMTP] SMTP connection failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "500",
			"message": fmt.Sprintf("SMTP connection failed: %v", err),
			"success": false,
			"debug": fiber.Map{
				"server":   cfg.MailServer,
				"port":     cfg.MailPort,
				"username": cfg.MailUsername,
				"error":    err.Error(),
			},
		})
	}

	log.Println("[TestSMTP] SMTP connection successful!")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"code":    "200",
		"message": "SMTP connection successful!",
		"success": true,
		"debug": fiber.Map{
			"server":   cfg.MailServer,
			"port":     cfg.MailPort,
			"username": cfg.MailUsername,
			"note":     "Test email would be sent to test@example.com if not for security reasons",
		},
	})
}
