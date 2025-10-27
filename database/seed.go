package database

import (
	"log"
	"os"
	"time"

	"autentikasi/models"
	"autentikasi/utils"
)

// SeedDev creates a developer user and a few sample transactions when
// the environment variable DEV_SEED is set to "true". It's guarded so it
// won't overwrite existing users or transactions.
func SeedDev() error {
	if os.Getenv("DEV_SEED") != "true" {
		return nil
	}

	if DB == nil {
		log.Println("[SEED] DB is not initialized; skipping seed")
		return nil
	}

	// Check if dev user already exists
	var existing models.User
	if err := DB.First(&existing, "email = ?", "dev@example.com").Error; err == nil {
		// user exists; check if any transactions exist for them
		var cnt int64
		DB.Model(&models.Transaction{}).Where("user_id = ?", existing.ID).Count(&cnt)
		if cnt > 0 {
			log.Println("[SEED] Dev user and transactions already present; skipping seeding")
			return nil
		}
	}

	// create dev user
	passHash, _ := utils.Hash("password")
	u := models.User{
		Nama:     "Dev User",
		Email:    "dev@example.com",
		Password: passHash,
	}
	if err := DB.Create(&u).Error; err != nil {
		log.Printf("[SEED] failed to create dev user: %v", err)
		return err
	}
	log.Printf("[SEED] created dev user id=%d email=%s", u.ID, u.Email)

	// create sample transactions for the dev user
	now := time.Now()
	txs := []models.Transaction{
		{
			UserID:     u.ID,
			Jenis:      "pemasukan",
			Kategori:   "Gaji",
			Jumlah:     4500000,
			Metode:     "Transfer",
			Keterangan: "Gaji Bulanan",
			CreatedAt:  &now,
		},
		{
			UserID:     u.ID,
			Jenis:      "pengeluaran",
			Kategori:   "Makan malam",
			Jumlah:     20000,
			Metode:     "Tunai",
			Keterangan: "Makan Soto",
			CreatedAt:  &now,
		},
		{
			UserID:     u.ID,
			Jenis:      "pengeluaran",
			Kategori:   "Transport",
			Jumlah:     12000,
			Metode:     "Tunai",
			Keterangan: "BBM motor",
			CreatedAt:  &now,
		},
	}

	if err := DB.Create(&txs).Error; err != nil {
		log.Printf("[SEED] failed to create transactions: %v", err)
		return err
	}
	log.Printf("[SEED] created %d sample transactions for dev user", len(txs))
	return nil
}
