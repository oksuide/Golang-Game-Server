package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gameCore/config"
	"gameCore/pkg/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	DB    *gorm.DB
	sqlDB *sql.DB
)

// Connecting to a database
func Connect(dbConfig config.DatabaseConfig) error {
	var err error

	// Формируем DNS строку из конфига
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		dbConfig.Host,
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Name,
		dbConfig.Port,
		dbConfig.SSLMode,
	)

	// Setup the GORM logger
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}

	// Establishing a connection
	DB, err = gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get the underlying sql.DB to configure the pool
	sqlDB, err = DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// Configuring the connection pool
	sqlDB.SetMaxOpenConns(dbConfig.MaxOpenConns)
	sqlDB.SetMaxIdleConns(dbConfig.MaxIdleConns)

	// Checking connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	fmt.Println("✅ Database connection established")
	return nil
}

func InitTables() error {
	err := DB.AutoMigrate(
		&models.User{},
		&models.GameSession{},
		&models.Leaderboard{},
		&models.Matchmaking{},
		&models.ChatMessage{},
		&models.Player{},
	)
	if err != nil {
		return fmt.Errorf("error migrating database: %w", err)
	}

	// Добавляем индексы, которые GORM не создаёт автоматически
	if err := createAdditionalIndexes(); err != nil {
		return err
	}

	fmt.Println("Database migration completed successfully.")
	return nil
}

// Custom index
func createAdditionalIndexes() error {
	// Пример создания составного индекса
	if err := DB.Exec(
		"CREATE INDEX IF NOT EXISTS idx_matchmaking_user_status ON matchmakings(user_id, status)",
	).Error; err != nil {
		return fmt.Errorf("failed to create matchmaking index: %w", err)
	}

	// Индекс для часто используемых запросов
	if err := DB.Exec(
		"CREATE INDEX IF NOT EXISTS idx_chat_messages_session ON chat_messages(game_session_id, created_at)",
	).Error; err != nil {
		return fmt.Errorf("failed to create chat messages index: %w", err)
	}

	return nil
}

// Graceful shutdown
func Close() error {
	if sqlDB != nil {
		if err := sqlDB.Close(); err != nil {
			return fmt.Errorf("failed to close database: %w", err)
		}
		fmt.Println("🗄️ Database connection closed")
	}
	return nil
}
