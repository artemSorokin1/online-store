package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/artemSorokin1/products-grpc-api/internal/db"
	"github.com/artemSorokin1/products-grpc-api/internal/kafka"
	"github.com/artemSorokin1/products-grpc-api/internal/repository"
	"github.com/artemSorokin1/products-grpc-api/pkg/api"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Инициализация базы данных
	database, err := db.NewPostgresDB()
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}
	defer database.Close()

	// Инициализация Kafka producer
	kafkaBrokers := []string{getEnv("KAFKA_BROKER", "localhost:9092")}
	kafkaTopic := getEnv("KAFKA_TOPIC", "products")

	producer, err := kafka.NewProducer(kafkaBrokers, kafkaTopic)
	if err != nil {
		log.Fatalf("Ошибка создания Kafka producer: %v", err)
	}
	defer producer.Close()

	// Инициализация репозитория
	repo := repository.NewProductRepository(database, producer)

	// Инициализация и запуск consumer'а
	consumerGroupID := getEnv("KAFKA_CONSUMER_GROUP", "products-service")
	consumer, err := kafka.NewConsumer(kafkaBrokers, consumerGroupID, kafkaTopic, kafka.NewProductEventHandler(repo))
	if err != nil {
		log.Fatalf("Ошибка создания Kafka consumer: %v", err)
	}
	defer consumer.Close()

	// Запуск consumer'а в отдельной горутине
	go func() {
		if err := consumer.Start(ctx); err != nil {
			log.Printf("Ошибка работы consumer'а: %v", err)
			cancel()
		}
	}()

	// Запуск gRPC сервера в отдельной горутине
	go func() {
		if err := api.RunProductsServer(":8081", repo); err != nil {
			log.Printf("Ошибка работы gRPC сервера: %v", err)
			cancel()
		}
	}()

	// Ожидание сигнала завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Получен сигнал завершения, останавливаем сервис...")
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
