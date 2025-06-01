package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"searchservice/internal/handler"
	"searchservice/internal/repository"
	"searchservice/internal/service"
)

func main() {
	// 1. Чтение переменных окружения
	kafkaBroker := os.Getenv("KAFKA_BROKER")
	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	groupID := os.Getenv("KAFKA_GROUP_ID")
	esURL := os.Getenv("ES_URL")

	if kafkaBroker == "" || kafkaTopic == "" || groupID == "" || esURL == "" {
		log.Fatal("Environment variables KAFKA_BROKER, KAFKA_TOPIC, KAFKA_GROUP_ID and ES_URL must be set")
	}

	// 2. Создаём репозиторий Elasticsearch
	esRepo, err := repository.NewElasticRepository(esURL)
	if err != nil {
		log.Fatalf("Failed to create Elasticsearch repository: %s", err)
	}

	// 3. Создаём бизнес-сервис
	prodService := service.NewProductService(esRepo)

	// 4. Запускаем Kafka-консьюмер в отдельной горутине
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	kafkaConsumer := repository.NewKafkaConsumer(
		kafkaBroker,
		kafkaTopic,
		groupID,
		prodService.HandleNewProduct, // callback
	)

	go func() {
		if err := kafkaConsumer.Start(ctx); err != nil {
			log.Printf("Kafka consumer stopped with error: %s", err)
		}
	}()

	// 5. Настраиваем HTTP-сервер и роутинг
	searchHandler := handler.NewSearchHandler(prodService)
	mux := http.NewServeMux()
	mux.Handle("/search", searchHandler)

	server := &http.Server{
		Addr:         ":8085",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// 6. Горутина для graceful shutdown
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-shutdownCh
		log.Println("Shutting down HTTP server...")
		server.Close()
		cancel() // остановим Kafka-консьюмер
	}()

	log.Println("HTTP server listening on :8080")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("HTTP server error: %s", err)
	}
	log.Println("Server stopped")
}
