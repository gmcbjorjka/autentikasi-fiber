package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Nama      string         `gorm:"size:100;not null;column:nama" json:"nama"`
	Email     string         `gorm:"size:100;not null;uniqueIndex;column:email" json:"email"`
	Password  string         `gorm:"size:255;not null;column:password" json:"-"`
	Pin       *string        `gorm:"size:255;column:pin" json:"-"`                     // opsional
	ImgURL    *string        `gorm:"size:255;column:img_url" json:"img_url,omitempty"` // opsional
	Phone     *string        `gorm:"size:32;column:phone" json:"phone,omitempty"`
	Gender    *string        `gorm:"size:32;column:gender" json:"gender,omitempty"`
	Birthday  *time.Time     `gorm:"column:birthday" json:"birthday,omitempty"`
	Status    *string        `gorm:"size:32;column:status" json:"status,omitempty"`
	CreatedAt *time.Time     `gorm:"column:created_at" json:"created_at,omitempty"`
	UpdatedAt *time.Time     `gorm:"column:updated_at" json:"updated_at,omitempty"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (User) TableName() string { return "users" }
