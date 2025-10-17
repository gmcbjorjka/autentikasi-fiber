package models

import (
	"time"

	"gorm.io/gorm"
)

type Transaction struct {
	ID         uint64         `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UserID     uint64         `gorm:"not null;column:user_id" json:"user_id"`
	Jenis      string         `gorm:"size:50;not null;column:jenis" json:"jenis"` // contoh: salary, grocery
	Kategori   string         `gorm:"size:100;column:kategori" json:"kategori"`
	Jumlah     float64        `gorm:"column:jumlah;not null" json:"jumlah"`
	Metode     string         `gorm:"size:100;column:metode" json:"metode"`
	Keterangan string         `gorm:"size:255;column:keterangan" json:"keterangan"`
	CreatedAt  *time.Time     `gorm:"column:created_at" json:"created_at,omitempty"`
	UpdatedAt  *time.Time     `gorm:"column:updated_at" json:"updated_at,omitempty"`
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (Transaction) TableName() string { return "transactions" }
