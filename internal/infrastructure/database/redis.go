package database

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/tharunn0/E-Commerce-Go/internal/config"
	"go.uber.org/zap"
)

func InitRedis(ctx context.Context, rediscfg config.RedisSettings, log *zap.Logger) *redis.Client {

	rdb := redis.NewClient(&redis.Options{
		Addr:         rediscfg.URL,
		Password:     rediscfg.Password,
		DB:           rediscfg.DB,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		DialTimeout:  5 * time.Second,
	})

	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := rdb.Ping(timeoutCtx).Err(); err != nil {
		log.Fatal("REDIS_CONNECTION_FAILED", zap.Error(err))
	}

	log.Info("REDIS_CONNECTED_SUCCESSFULLY")

	return rdb
}
