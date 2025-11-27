package models

import (
	"time"

	"gorm.io/gorm"
)

type PasswordReset struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Email     string         `gorm:"size:100;not null;index;column:email" json:"email"`
	OTP       string         `gorm:"size:10;not null;column:otp" json:"otp"`
	ExpiresAt time.Time      `gorm:"column:expires_at;index" json:"expires_at"`
	CreatedAt *time.Time     `gorm:"column:created_at" json:"created_at,omitempty"`
	UpdatedAt *time.Time     `gorm:"column:updated_at" json:"updated_at,omitempty"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (PasswordReset) TableName() string { return "password_resets" }
