package main

import (
	"comment_service/internal/domain"
	"comment_service/internal/handler"
	"comment_service/internal/repository"
	"comment_service/internal/service"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func waitForDB(host string, port int) error {
	timeout := time.After(30 * time.Second)
	tick := time.Tick(1 * time.Second)

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for database")
		case <-tick:
			conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
			if err == nil {
				conn.Close()
				return nil
			}
		}
	}
}

func main() {
	// Ждем готовности базы данных
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "comment_db"
	}
	if err := waitForDB(dbHost, 5432); err != nil {
		log.Fatal("Failed to wait for database:", err)
	}

	// Подключение к базе данных
	dsn := fmt.Sprintf("host=%s user=root password=123 dbname=comments port=5432 sslmode=disable", dbHost)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Автоматическая миграция схемы
	err = db.AutoMigrate(&domain.Comment{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Инициализация зависимостей
	commentRepo := repository.NewCommentRepository(db)
	commentService := service.NewCommentService(commentRepo)
	commentHandler := handler.NewCommentHandler(commentService)

	// Настройка маршрутов
	r := gin.Default()

	// Группа маршрутов для комментариев
	comments := r.Group("/api/comments")
	{
		comments.POST("", commentHandler.CreateComment)
		comments.GET("/product/:product_id", commentHandler.GetProductComments)
		comments.DELETE("/:id", commentHandler.DeleteComment)
	}

	// Запуск сервера в горутине
	go func() {
		if err := r.Run(":8084"); err != nil {
			log.Fatal("Failed to start server:", err)
		}
	}()

	// Ожидание сигнала для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
}
