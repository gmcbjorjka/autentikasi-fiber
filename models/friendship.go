package models

import (
	"time"

	"gorm.io/gorm"
)

type Friendship struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UserID    uint64         `gorm:"not null;column:user_id" json:"user_id"`
	FriendID  uint64         `gorm:"not null;column:friend_id" json:"friend_id"`
	Status    string         `gorm:"size:20;not null;column:status;default:pending" json:"status"` // pending, accepted, rejected
	CreatedAt *time.Time     `gorm:"column:created_at" json:"created_at,omitempty"`
	UpdatedAt *time.Time     `gorm:"column:updated_at" json:"updated_at,omitempty"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (Friendship) TableName() string { return "friendships" }

// BeforeCreate hook to ensure UserID < FriendID for uniqueness
func (f *Friendship) BeforeCreate(tx *gorm.DB) (err error) {
	if f.UserID > f.FriendID {
		f.UserID, f.FriendID = f.FriendID, f.UserID
	}
	return
}
