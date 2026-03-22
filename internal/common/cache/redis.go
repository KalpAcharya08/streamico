package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RedisConfig holds connection settings for Redis
type RedisConfig struct {
	Host     string
	Port     string
	Password string
}

func NewRedisClient(ctx context.Context, cfg RedisConfig, logger *zap.Logger) (*redis.Client, error) {
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	logger.Info("connecting to Redis", zap.String("addr", addr))

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Password,
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	logger.Info("✅ connected to Redis successfully")
	return client, nil
}
