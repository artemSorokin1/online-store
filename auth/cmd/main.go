package main

import (
	"auth/internal/config"
	"auth/internal/routes"
	"auth/pkg/server"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strings"

	api "auth/internal/grpc_server"

	"github.com/jmoiron/sqlx"
)

func main() {
	cfg := config.MustLoad()

	err := runMigrationsWithSqlx(cfg)
	if err != nil {
		log.Fatal("migration failed: ", err)
	}

	router, db, err := routes.SetupApp(cfg)
	if err != nil {
		log.Fatal("cant setup app", err)
	}
	defer db.DB.Close()

	go server.MustRun(router, cfg)

	go api.RunSellersServer(cfg)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch
	slog.Error("shutting down server...")
}

func runMigrationsWithSqlx(cfg *config.Config) error {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBCfg.Host,
		cfg.DBCfg.Port,
		cfg.DBCfg.Username,
		cfg.DBCfg.Password,
		cfg.DBCfg.DBName,
	)
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	files, err := os.ReadDir("./migrations")
	if err != nil {
		return err
	}
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".up.sql") {
			content, err := os.ReadFile("./migrations/" + file.Name())
			if err != nil {
				return err
			}
			if _, err := db.Exec(string(content)); err != nil {
				return err
			}
		}
	}
	return nil
}
