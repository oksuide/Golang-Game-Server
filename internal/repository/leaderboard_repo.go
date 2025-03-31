package repository

import (
	"context"
	"gameCore/pkg/models"

	"gorm.io/gorm"
)

type LeaderboardRepo struct {
	DB *gorm.DB
}

func NewLeaderboardRepo(db *gorm.DB) *LeaderboardRepo {
	return &LeaderboardRepo{DB: db}
}

func (r *LeaderboardRepo) LeaderboardExists(ctx context.Context, userID uint) (bool, error) {
	var count int64
	err := r.DB.WithContext(ctx).
		Model(&models.Leaderboard{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count > 0, err
}

// Создает новую запись в таблице leaderboard
func (r *LeaderboardRepo) CreateLeaderboard(ctx context.Context, leaderboard *models.Leaderboard) error {
	return r.DB.WithContext(ctx).Create(leaderboard).Error
}

// Получает запись в таблице leaderboard по userID
func (r *LeaderboardRepo) GetLeaderboard(ctx context.Context, userID uint) (*models.Leaderboard, error) {
	var leaderboard models.Leaderboard
	err := r.DB.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&leaderboard).Error
	if err != nil {
		return nil, err
	}
	return &leaderboard, nil
}

type LeaderboardRepository interface {
	LeaderboardExists(ctx context.Context, userID uint) (bool, error)
	CreateLeaderboard(ctx context.Context, leaderboard *models.Leaderboard) error
	GetLeaderboard(ctx context.Context, userID uint) (*models.Leaderboard, error)
}
