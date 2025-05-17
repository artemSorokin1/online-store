package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
)

type Config struct {
	ServerCfg ServerConfig `yaml:"server"`
	DBCfg     DBConfig     `yaml:"storage"`
}

type DBConfig struct {
	Host     string `yaml:"DATABASE_HOST" env:"DATABASE_HOST" env-default:"localhost"`
	Password string `yaml:"DATABASE_PASSWORD" env:"DATABASE_PASSWORD"`
	Username string `yaml:"DATABASE_USERNAME" env:"DATABASE_USERNAME"`
	DBName   string `yaml:"DATABASE_NAME" env:"DATABASE_NAME" env-default:"delivery"`
	Port     string `yaml:"DATABASE_PORT" env:"DATABASE_PORT" env-default:"5433"`
}

type ServerConfig struct {
	Port string `yaml:"PORT"`
}

func New() *Config {
	var cfg Config
	PathToConfig := os.Getenv("CONFIG_PATH")
	fmt.Println(PathToConfig)
	err := cleanenv.ReadConfig(PathToConfig, &cfg)
	if err != nil {
		panic(err)
	}

	return &cfg
}
