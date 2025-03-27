package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `gorm:"type:varchar(50);not null;unique"`
	Email    string `gorm:"type:text;not null;unique"`
	Password string `gorm:"type:text;not null"`
	Sessions []SessionToken
}

type SessionToken struct {
	gorm.Model
	UserID    uint   `gorm:"not null"`
	Token     string `gorm:"type:text;not null;unique"`
	ExpiresAt time.Time
}
