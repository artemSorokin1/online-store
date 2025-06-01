package main

import (
	"auth_service/internal/config"
	"auth_service/internal/repositiry/storage"
	"auth_service/internal/transport/grpc"
	"auth_service/pkg/logger"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"go.uber.org/zap"
)

//const (
//	CntGenUsers = 5_000_000
//)

func main() {
	cfg := config.New()

	env := os.Getenv("ENV")
	if env == "" {
		slog.Error("ENV variable is not set")
		os.Exit(1)
	}
	mainLogger, err := logger.InitLogger(env)
	if err != nil {
		slog.Error("failed to initialize logger", err)
	}

	stor, err := storage.New(cfg, mainLogger)
	if err != nil {
		mainLogger.Fatal("failed to create storage", zap.Error(err))
	}

	fmt.Println(cfg)

	grpcServer := grpc.New(cfg.ServerCfg, stor, mainLogger)
	go grpcServer.MustStart()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	sign := <-ch

	mainLogger.Error("shutting down server", zap.String("signal", sign.String()))
	grpcServer.GracefulStop()

}
