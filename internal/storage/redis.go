package storage

import (
	"context"
	"fmt"
	"gameCore/internal/config"
	"time"

	"github.com/go-redis/redis"
)

var RedisClient *redis.Client

func InitRedis(cfg config.RedisConfig) error {
	// Формируем адрес подключения
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	_, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := RedisClient.Ping().Err(); err != nil {
		return fmt.Errorf("redis connection failed: %w", err)
	}

	fmt.Println("✅ Redis connection established")
	return nil
}
