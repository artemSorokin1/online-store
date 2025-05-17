package client

import (
	"context"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"notification_service/internal/config"
	"notification_service/internal/notifiction"
)

type RedisSubscriber struct {
	client *redis.Client
	config *config.RedisConfig
}

func NewRedisClient(cfg *config.RedisConfig) *RedisSubscriber {
	return &RedisSubscriber{
		client: redis.NewClient(&redis.Options{
			Addr: cfg.Host + ":" + cfg.Port,
			DB:   1,
		}),
		config: cfg,
	}
}

func (r *RedisSubscriber) ListenAndServe(emailSenderCfg *config.NotifyConfig) {
	emailSeneder := notifiction.New(emailSenderCfg)

	sub := r.client.Subscribe(context.Background(), r.config.ChanelName)
	defer sub.Close()

	ch := sub.Channel()

	for msg := range ch {
		slog.Info("payment message received", slog.String("message", msg.Payload))
		go notifiction.NotifyUser(emailSeneder, msg.String())
	}
}
