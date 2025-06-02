package main

import (
	"Web-shop/products_service/pkg/api"
	"log"
)

func main() {
	if err := api.RunProductsServer("0.0.0.0:50051"); err != nil {
		log.Fatalf("Failed to start ProductsService: %v", err)
	}
}
