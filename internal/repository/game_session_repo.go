package repository

import (
	"context"
	"gameCore/pkg/models"

	"gorm.io/gorm"
)

type GameSessionRepo struct {
	DB *gorm.DB
}

func NewGameSessionRepo(db *gorm.DB) *GameSessionRepo {
	return &GameSessionRepo{DB: db}
}

func (r *GameSessionRepo) GameSessionExists(ctx context.Context, gameID string) (bool, error) {
	var count int64
	err := r.DB.WithContext(ctx).
		Model(&models.GameSession{}).
		Where("game_id = ?", gameID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *GameSessionRepo) CreateGameSession(ctx context.Context, gameSession *models.GameSession) error {
	return r.DB.WithContext(ctx).Create(gameSession).Error
}

func (r *GameSessionRepo) GetGameSession(ctx context.Context, gameID string) (*models.GameSession, error) {
	var gameSession models.GameSession
	err := r.DB.WithContext(ctx).
		Where("game_id = ?", gameID).
		First(&gameSession).Error
	if err != nil {
		return nil, err
	}
	return &gameSession, nil
}

func (r *GameSessionRepo) GetAllGameSessions(ctx context.Context) ([]models.GameSession, error) {
	var gameSessions []models.GameSession
	err := r.DB.WithContext(ctx).
		Find(&gameSessions).Error
	if err != nil {
		return nil, err
	}
	return gameSessions, nil
}

func (r *GameSessionRepo) UpdateGameSession(ctx context.Context, gameSession *models.GameSession) error {
	return r.DB.WithContext(ctx).Save(gameSession).Error
}

type GameSessionRepository interface {
	GameSessionExists(ctx context.Context, gameID string) (bool, error)
	CreateGameSession(ctx context.Context, gameSession *models.GameSession) error
	GetGameSession(ctx context.Context, gameID string) (*models.GameSession, error)
	GetAllGameSessions(ctx context.Context) ([]models.GameSession, error)
	UpdateGameSession(ctx context.Context, gameSession *models.GameSession) error
}
