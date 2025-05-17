package postgres

import (
	"auth_service/internal/config"
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type DB struct {
	Db *sqlx.DB
}

func New(config *config.Config) (*DB, error) {
	cfg := config.StorageCfg

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", cfg.Host, cfg.Username, cfg.Password, cfg.DBName, cfg.Port)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to db: %w", err)
	}

	if _, err := db.Conn(context.Background()); err != nil {
		return nil, fmt.Errorf("unable to connect to db: %w", err)
	}

	return &DB{Db: db}, nil
}
