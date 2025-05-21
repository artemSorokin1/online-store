package inmem

import (
	"context"
	"dlivery_service/delivery_service/internal/config"
	"dlivery_service/delivery_service/internal/models"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/redis/go-redis/v9"
)

type RedisClientForNotify struct {
	cfg    *config.RedisConfig
	client *redis.Client
}

func NewRedisClientForNotify(cfg *config.RedisConfig) *RedisClientForNotify {
	client := redis.NewClient(&redis.Options{
		Addr: cfg.Host + ":" + cfg.Port,
		DB:   1,
	})

	return &RedisClientForNotify{
		cfg:    cfg,
		client: client,
	}
}

func (r *RedisClientForNotify) Publish(email string) {
	msg := fmt.Sprintf("payment success email: %s", email)
	err := r.client.Publish(context.Background(), r.cfg.ChanelName, msg).Err()
	if err != nil {
		slog.Error("Error publishing message", slog.String("error", err.Error()))
		return
	}
}

type RedisClientForCart struct {
	cfg    *config.RedisConfig
	client *redis.Client
}

func NewRedisClientForCart(cfg *config.RedisConfig) *RedisClientForCart {
	client := redis.NewClient(&redis.Options{
		Addr: cfg.Host + ":" + cfg.Port,
		DB:   2,
	})

	return &RedisClientForCart{
		cfg:    cfg,
		client: client,
	}
}

func (r *RedisClientForCart) SaveCart(userID int64, cart []models.CartItem) error {
	jsonCartString, err := json.Marshal(cart)
	if err != nil {
		slog.Error("Error marshalling cart to JSON", slog.String("error", err.Error()))
	}
	err = r.client.Set(context.Background(), strconv.Itoa(int(userID)), jsonCartString, 10).Err()
	if err != nil {
		slog.Error("Error saving cart to Redis", slog.String("error", err.Error()))
		return err
	}
	slog.Info("Cart saved to Redis", slog.String("userID", strconv.Itoa(int(userID))))
	return nil
}

func (r *RedisClientForCart) GetCart(userID int64) ([]models.CartItem, error) {
	key := strconv.FormatInt(userID, 10)
	val, err := r.client.Get(context.Background(), key).Result()
	if err != nil {
		if err == redis.Nil {
			return []models.CartItem{}, nil
		}
		slog.Error("Error getting cart from Redis", slog.String("error", err.Error()))
		return nil, err
	}

	var cart []models.CartItem
	if err := json.Unmarshal([]byte(val), &cart); err != nil {
		slog.Error("Error unmarshalling cart from JSON", slog.String("error", err.Error()))
		return nil, err
	}

	return cart, nil
}
