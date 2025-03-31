package repository

import (
	"context"
	"gameCore/pkg/models"

	"gorm.io/gorm"
)

type ChatMessageRepo struct {
	DB *gorm.DB
}

func NewChatMessageRepo(db *gorm.DB) *ChatMessageRepo {
	return &ChatMessageRepo{DB: db}
}

// Проверка, существует ли сообщение с данным ID
func (r *ChatMessageRepo) ChatMessageExists(ctx context.Context, messageID uint) (bool, error) {
	var count int64
	err := r.DB.WithContext(ctx).
		Model(&models.ChatMessage{}).
		Where("id = ?", messageID).
		Count(&count).Error
	return count > 0, err
}

// Создание нового сообщения в чате
func (r *ChatMessageRepo) CreateChatMessage(ctx context.Context, chatMessage *models.ChatMessage) error {
	return r.DB.WithContext(ctx).Create(chatMessage).Error
}

// Получение сообщения по ID
func (r *ChatMessageRepo) GetChatMessage(ctx context.Context, messageID uint) (*models.ChatMessage, error) {
	var chatMessage models.ChatMessage
	err := r.DB.WithContext(ctx).
		Where("id = ?", messageID).
		First(&chatMessage).Error
	if err != nil {
		return nil, err
	}
	return &chatMessage, nil
}

// Получение всех сообщений по GameSessionID (для загрузки истории чата)
func (r *ChatMessageRepo) GetMessagesBySession(ctx context.Context, gameSessionID uint) ([]models.ChatMessage, error) {
	var messages []models.ChatMessage
	err := r.DB.WithContext(ctx).
		Where("game_session_id = ?", gameSessionID).
		Order("created_at ASC").
		Find(&messages).Error
	if err != nil {
		return nil, err
	}
	return messages, nil
}

// Получение всех сообщений пользователя (например, для поиска его активности в чате)
func (r *ChatMessageRepo) GetMessagesByUser(ctx context.Context, userID uint) ([]models.ChatMessage, error) {
	var messages []models.ChatMessage
	err := r.DB.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at ASC").
		Find(&messages).Error
	if err != nil {
		return nil, err
	}
	return messages, nil
}

type ChatMessageRepository interface {
	ChatMessageExists(ctx context.Context, messageID uint) (bool, error)
	CreateChatMessage(ctx context.Context, chatMessage *models.ChatMessage) error
	GetChatMessage(ctx context.Context, messageID uint) (*models.ChatMessage, error)
	GetMessagesBySession(ctx context.Context, gameSessionID uint) ([]models.ChatMessage, error)
	GetMessagesByUser(ctx context.Context, userID uint) ([]models.ChatMessage, error)
}
