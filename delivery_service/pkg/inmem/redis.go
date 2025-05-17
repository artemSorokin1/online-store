package inmem

import (
	"context"
	"dlivery_service/delivery_service/internal/config"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log/slog"
)

type RedisClient struct {
	cfg    *config.RedisConfig
	client *redis.Client
}

func NewRedisClient(cfg *config.RedisConfig) *RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr: cfg.Host + ":" + cfg.Port,
		DB:   1,
	})

	return &RedisClient{
		cfg:    cfg,
		client: client,
	}
}

func (r *RedisClient) Publish(email string) {
	msg := fmt.Sprintf("payment success email: %s", email)
	err := r.client.Publish(context.Background(), r.cfg.ChanelName, msg).Err()
	if err != nil {
		slog.Error("Error publishing message", slog.String("error", err.Error()))
		return
	}
}
