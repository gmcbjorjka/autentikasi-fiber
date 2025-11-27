package database

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"autentikasi/config"
	"autentikasi/models"
)

var DB *gorm.DB

func Connect(cfg *config.Config) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&loc=Local",
		cfg.MySQLUser, cfg.MySQLPass, cfg.MySQLHost, cfg.MySQLPort, cfg.MySQLDB,
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	if err := db.AutoMigrate(&models.User{}, &models.Transaction{}, &models.Friendship{}, &models.PasswordReset{}, &models.AccountHistory{}); err != nil {
		return err
	}

	DB = db
	log.Println("📦 AutoMigrate complete")

	// backfill phone_digits for existing users (safe operation)
	go func() {
		var users []models.User
		if err := DB.Find(&users).Error; err != nil {
			log.Printf("[DB] failed to fetch users for phone_digits backfill: %v", err)
			return
		}
		for _, u := range users {
			if u.Phone == nil || *u.Phone == "" {
				continue
			}
			// compute digits-only
			var b []rune
			for _, r := range *u.Phone {
				if r >= '0' && r <= '9' {
					b = append(b, r)
				}
			}
			s := string(b)
			if s == "" {
				continue
			}
			// update if different or empty
			if u.PhoneDigits == nil || *u.PhoneDigits != s {
				DB.Model(&u).Update("phone_digits", s)
			}
		}
		log.Println("[DB] phone_digits backfill completed")
	}()

	return nil
}
