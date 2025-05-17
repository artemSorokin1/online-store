package main

import (
	"auth_service/internal/config"
	"auth_service/internal/repositiry/storage"
	"auth_service/internal/transport/grpc"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
)

//const (
//	CntGenUsers = 5_000_000
//)

func main() {
	cfg := config.New()

	stor, err := storage.New(cfg)
	if err != nil {
		slog.Error("storage is not ready", err)
		os.Exit(1)
	}

	fmt.Println(cfg)

	grpcServer := grpc.New(cfg.ServerCfg, stor)
	//go utils.GenUsers(CntGenUsers, stor)
	go grpcServer.MustStart()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	sign := <-ch

	slog.Info("Got signal: ", sign)
	grpcServer.GracefulStop()

}
