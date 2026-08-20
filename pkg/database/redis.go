package database

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/redis/go-redis/v9"

	"github.com/team-task-api/internal/config"
)

func NewRedis(cfg config.RedisConfig) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	slog.Info("connected to redis", "addr", cfg.Addr)
	return rdb, nil
}
