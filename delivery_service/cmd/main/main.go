package main

import (
	"context"
	"dlivery_service/delivery_service/internal/config"
	"dlivery_service/delivery_service/internal/repository/storage"
	"dlivery_service/delivery_service/pkg/logger"
	"dlivery_service/delivery_service/pkg/server"
	"log/slog"
	"os"
	"os/signal"
)

func main() {
	ctx := context.Background()

	env := os.Getenv("ENV")

	mainLogger, err := logger.InitLogger(env)
	if err != nil {
		slog.Error("failed to initialize logger", err)
	}

	cfg := config.New()

	stor, err := storage.New(cfg, mainLogger)
	if err != nil {
		panic(err)
	}

	serv := server.New(ctx, stor, mainLogger)

	go serv.MustRun(cfg)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill)

	<-sigChan

	mainLogger.Error("shutting down server")

}
