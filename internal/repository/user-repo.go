package repository

import (
	"context"
	"fmt"
	"gameCore/pkg/models"

	"gorm.io/gorm"
)

type UserRepo struct {
	DB *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{DB: db}
}

func (r *UserRepo) UserExists(ctx context.Context, username string) (bool, error) {
	var count int64
	if err := r.DB.WithContext(ctx).
		Model(&models.User{}).
		Where("username = ?", username).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *UserRepo) CreateUser(ctx context.Context, user *models.User) error {
	return r.DB.WithContext(ctx).Create(user).Error
}

func (r *UserRepo) GetUser(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	if err := r.DB.WithContext(ctx).
		Where("username = ?", username).
		First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) EmailExists(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.DB.Model(&models.User{}).
		Where("email = ?", email).
		Count(&count).
		Error

	if err != nil {
		return false, fmt.Errorf("email exists check failed: %w", err)
	}
	return count > 0, nil
}

type UserRepository interface {
	UserExists(ctx context.Context, username string) (bool, error)
	CreateUser(ctx context.Context, user *models.User) error
	GetUser(ctx context.Context, username string) (*models.User, error)
	EmailExists(ctx context.Context, email string) (bool, error)
}
