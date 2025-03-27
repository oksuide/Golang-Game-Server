package models

import (
	"time"

	"gorm.io/gorm"
)

// Player in session
type Player struct {
	gorm.Model
	UserID        uint `gorm:"not null"`
	GameSessionID uint `gorm:"not null"`
	Score         int  `gorm:"default:0"` // score in currect match
	IsReady       bool `gorm:"default:false"`
}

// Lobby
type GameSession struct {
	gorm.Model
	GameID     string    `gorm:"type:varchar(50);not null;unique"`
	StartTime  time.Time `gorm:"not null"`
	EndTime    time.Time
	MaxPlayers int      `gorm:"not null"`
	Players    []Player `gorm:"foreignKey:GameSessionID"`
}

// Queue for the match
type Matchmaking struct {
	gorm.Model
	UserID   uint      `gorm:"not null"`
	Rank     int       `gorm:"not null"`
	Status   string    `gorm:"type:varchar(20);default:'searching'"` // searching, matched, canceled
	JoinedAt time.Time `gorm:"autoCreateTime"`
}
