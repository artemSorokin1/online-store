package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log/slog"
	"os"
)

type RedisConfig struct {
	Host       string `yml:"host"`
	Port       string `yml:"port"`
	ChanelName string `yml:"chanel_name"`
}

func NewRedisConfig() *RedisConfig {
	pathToConfig := os.Getenv("REDIS_CONFIG_PATH")
	var cfg RedisConfig

	if pathToConfig == "" {
		slog.Error("Path to redis config is not set")
		os.Exit(1)
	}

	err := cleanenv.ReadConfig(pathToConfig, &cfg)
	if err != nil {
		slog.Error("Error reading redis config", err)
		os.Exit(1)
	}

	return &cfg
}
