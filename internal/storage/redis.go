package storage

import (
	"context"
	"fmt"
	"gameCore/config"
	"time"

	"github.com/redis/go-redis/v9"
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

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := RedisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis connection failed: %w", err)
	}

	fmt.Println("✅ Redis connection established")
	return nil
}
