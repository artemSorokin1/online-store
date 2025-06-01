package main

import (
	"content_service/internal/api/product_client"
	"content_service/internal/api/seller_client"
	"content_service/internal/config"
	httpserver "content_service/internal/http_server"
	"log"
	"log/slog"
	"os"
	"os/signal"
)

func main() {
	// надо передать сюда адрес
	pcli, err := product_client.NewProductsClient()
	if err != nil {
		log.Fatalf("failed to create ProductsClient: %v", err)
	}

	scli, err := seller_client.NewSellersClient()
	if err != nil {
		log.Fatalf("failed to create ProductsClient: %v", err)
	}

	serverConfig := config.LoadConfig()

	server := httpserver.New(serverConfig, scli, pcli)

	go server.MustRun()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	sign := <-ch

	slog.Error("shutting down server with signal", slog.String("signal", sign.String()))

}
