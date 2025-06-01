package main

import (
	"log"

	"github.com/artemSorokin1/products-grpc-api/internal/db"
	"github.com/artemSorokin1/products-grpc-api/internal/repository"
	"github.com/artemSorokin1/products-grpc-api/pkg/api"
)

func main() {
	database, err := db.NewPostgresDB()
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}
	defer database.Close()

	repo := repository.NewProductRepository(database)

	if err := api.RunProductsServer(":8081", repo); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
