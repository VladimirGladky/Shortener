package redis

import (
	"context"
	"fmt"

	"github.com/wb-go/wbf/config"
	"github.com/wb-go/wbf/redis"
)

const URLCachePrefix = "shortener:url:"

func NewRedisClient(ctx context.Context, cfg *config.Config) (*redis.Client, error) {
	options := redis.Options{
		Address:   fmt.Sprintf("%s:%d", cfg.GetString("REDIS_HOST"), cfg.GetInt("REDIS_PORT")),
		Password:  cfg.GetString("REDIS_PASSWORD"),
		MaxMemory: cfg.GetString("REDIS_MAXMEMORY"),
		Policy:    "allkeys-lru",
	}

	client, err := redis.Connect(options)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}
	if err = client.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}
	return client, nil
}

func CacheKey(shortURL string) string {
	return URLCachePrefix + shortURL
}
