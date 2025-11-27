package models

import (
	"time"
)

type AccountHistory struct {
	ID          uint64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UserID      uint64     `gorm:"index;column:user_id" json:"user_id"`
	Event       string     `gorm:"size:64;column:event" json:"event"`
	Description string     `gorm:"size:255;column:description" json:"description"`
	CreatedAt   *time.Time `gorm:"column:created_at" json:"created_at"`
}

func (AccountHistory) TableName() string { return "account_histories" }
