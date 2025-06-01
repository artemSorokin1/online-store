package main

import (
	"log"
	"sellers_service/internal/api"
)

func main() {
	if err := api.RunSellersServer(":50052"); err != nil {
		log.Fatalf("Failed to start SellersService: %v", err)
	}
}
