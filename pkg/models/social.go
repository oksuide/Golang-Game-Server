package models

import (
	"time"

	"gorm.io/gorm"
)

// Chat in lobby
type ChatMessage struct {
	gorm.Model
	GameSessionID uint      `gorm:"not null"`
	UserID        uint      `gorm:"not null"`
	Message       string    `gorm:"type:text;not null"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
}

// Global leaderboard
type Leaderboard struct {
	gorm.Model
	UserID uint `gorm:"not null;unique"`
	Score  int  `gorm:"not null"` // global player score
	Rank   int  `gorm:"not null"`
}
