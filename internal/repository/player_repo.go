package repository

import (
	"context"
	"gameCore/pkg/models"

	"gorm.io/gorm"
)

type PlayerRepo struct {
	DB *gorm.DB
}

func NewPlayerRepo(db *gorm.DB) *PlayerRepo {
	return &PlayerRepo{DB: db}
}

// Проверка, существует ли игрок в сессии
func (r *PlayerRepo) PlayerExists(ctx context.Context, userID uint, gameSessionID uint) (bool, error) {
	var count int64
	err := r.DB.WithContext(ctx).
		Model(&models.Player{}).
		Where("user_id = ? AND game_session_id = ?", userID, gameSessionID).
		Count(&count).Error
	return count > 0, err
}

// Create new player
func (r *PlayerRepo) CreatePlayer(ctx context.Context, player *models.Player) error {
	return r.DB.WithContext(ctx).Create(player).Error
}

// Get player from id
func (r *PlayerRepo) GetPlayer(ctx context.Context, userID uint) (*models.Player, error) {
	var player models.Player
	err := r.DB.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&player).Error
	if err != nil {
		return nil, err
	}
	return &player, nil
}

// Получение всех игроков в сессии
func (r *PlayerRepo) GetPlayersBySession(ctx context.Context, gameSessionID uint) ([]models.Player, error) {
	var players []models.Player
	err := r.DB.WithContext(ctx).
		Where("game_session_id = ?", gameSessionID).
		Find(&players).Error
	if err != nil {
		return nil, err
	}
	return players, nil
}

type PlayerRepository interface {
	PlayerExists(ctx context.Context, userID uint) (bool, error)
	CreatePlayer(ctx context.Context, player *models.Player) error
	GetPlayer(ctx context.Context, userID uint) (*models.Player, error)
	GetPlayersBySession(ctx context.Context, gameSessionID uint) ([]models.Player, error)
}
