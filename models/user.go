package models

import "time"

type User struct {
	ID        uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	Nama      string     `gorm:"size:100;not null" json:"nama"`
	Email     string     `gorm:"size:100;unique;not null" json:"email"`
	Password  string     `gorm:"size:255;not null" json:"-"`
	Pin       string     `gorm:"size:255;not null" json:"-"`
	ImgUrl    *string    `gorm:"size:255" json:"img_url"`
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at"`
}
