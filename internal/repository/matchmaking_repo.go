package repository

import (
	"context"

	"gameCore/pkg/models"

	"gorm.io/gorm"
)

type MatchmakingRepository struct {
	db *gorm.DB
}

func NewMatchmakingRepository(db *gorm.DB) *MatchmakingRepository {
	return &MatchmakingRepository{db: db}
}

func (r *MatchmakingRepository) MatchmakingExists(ctx context.Context, playerID int) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Matchmaking{}).
		Where("player_id = ?", playerID).
		Count(&count).Error
	return count > 0, err
}

func (r *MatchmakingRepository) CreateMatchmaking(ctx context.Context, matchmaking *models.Matchmaking) error {
	return r.db.WithContext(ctx).Create(matchmaking).Error
}

func (r *MatchmakingRepository) GetMatchmaking(ctx context.Context, playerID int) (*models.Matchmaking, error) {
	var matchmaking models.Matchmaking
	err := r.db.WithContext(ctx).
		Where("player_id = ?", playerID).
		First(&matchmaking).Error
	if err != nil {
		return nil, err
	}
	return &matchmaking, nil
}
