package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log/slog"
	"os"
)

type NotifyConfig struct {
	EmailFrom string `yml:"email_from"`
	AppName   string `yml:"app_name"`
	Password  string `yml:"password"`
}

func NewNotifyConfig() *NotifyConfig {
	pathToConfig := os.Getenv("NOTIFY_CONFIG_PATH")
	var cfg NotifyConfig

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
