package config

import (
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type ServerConfig struct {
	Port string `yaml:"port"`
	Host string `yaml:"host"`
}

func LoadConfig() *ServerConfig {
	pathToConfig := os.Getenv("SERVER_CONFIG_PATH")

	if pathToConfig == "" {
		log.Fatal("SERVER_CONFIG_PATH environment variable is not set")
	}

	var cfg ServerConfig
	if err := cleanenv.ReadConfig(pathToConfig, &cfg); err != nil {
		log.Fatal("Failed to load server config: " + err.Error())
	}

	return &cfg
}
