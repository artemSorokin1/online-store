package inmem

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"os"
	"time"
)

type redisStorage struct {
	client *redis.Client
}

func (r *redisStorage) SaveToken(ctx context.Context, userId int64, token string) error {
	exp := os.Getenv("REFRESH_TOKEN_TTL")
	ttl, err := time.ParseDuration(exp)
	if err != nil {
		slog.Error("error parsing refresh token ttl", err)
		return err
	}
	r.client.Set(ctx, fmt.Sprintf("%d", userId), token, ttl)

	slog.Info("save token to redis", slog.Int64("userId", userId))

	return nil
}

func (r *redisStorage) GetToken(ctx context.Context, userId int64) (string, error) {
	res, err := r.client.Get(ctx, fmt.Sprintf("%d", userId)).Result()
	if err != nil {
		slog.Error("error getting token from redis", err)
		return "", err
	}
	if res == "" {
		slog.Warn("token not found in redis")
		return "", fmt.Errorf("token not found")
	}

	slog.Info("get token from redis", slog.Int64("userId", userId))

	return res, nil
}

func (r *redisStorage) RemoveToken(ctx context.Context, userId int64) error {
	err := r.client.Del(ctx, fmt.Sprintf("%d", userId)).Err()
	if err != nil {
		slog.Error("error deleting token from redis", err)
		return err
	}

	slog.Info("delete token from redis", slog.Int64("userId", userId))

	return nil
}

func NewRedisStorage() RefreshTokenStorage {
	host := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PORT")
	redisClient := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", host, port),
	})

	return &redisStorage{
		client: redisClient,
	}
}
