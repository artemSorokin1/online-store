package main

import (
	"context"
	"dlivery_service/delivery_service/internal/config"
	"dlivery_service/delivery_service/internal/repository/storage"
	"dlivery_service/delivery_service/pkg/logger"
	"dlivery_service/delivery_service/pkg/server"
	"github.com/labstack/gommon/log"
	"os"
	"os/signal"
)

const (
	ServiceName = "delivery_service"
	CntUsers    = 5_000_000
)

func main() {
	ctx := context.Background()
	mainLogger := logger.New(ServiceName)
	ctx = context.WithValue(ctx, logger.LoggerKey, mainLogger)

	cfg := config.New()

	stor, err := storage.New(cfg)
	if err != nil {
		panic(err)
	}

	serv := server.New(ctx, stor)

	go serv.MustRun(cfg)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill)

	<-sigChan

	log.Error("graceful shutdown")

}
