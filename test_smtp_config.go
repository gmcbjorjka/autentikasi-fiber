package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"

	"autentikasi/config"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	fmt.Println("=== SMTP Configuration ===")
	fmt.Printf("Server: %s\n", cfg.MailServer)
	fmt.Printf("Port: %s\n", cfg.MailPort)
	fmt.Printf("Username: %s\n", cfg.MailUsername)
	fmt.Printf("Password: %s (length: %d)\n", cfg.MailPassword, len(cfg.MailPassword))
	fmt.Printf("Sender: %s\n", cfg.MailDefaultSender)
	fmt.Println()

	// Test if we can parse the config
	if cfg.MailServer == "" {
		log.Fatal("MAIL_SERVER is empty!")
	}
	if cfg.MailUsername == "" {
		log.Fatal("MAIL_USERNAME is empty!")
	}
	if cfg.MailPassword == "" {
		log.Fatal("MAIL_PASSWORD is empty!")
	}

	fmt.Println("✅ All SMTP config loaded successfully!")
}
