package routes

import (
	"auth/internal/config"
	"auth/internal/handlers"
	"auth/internal/repository/postgres"
	log "auth/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func SetupApp(cfg *config.Config) (*gin.Engine, *sqlx.DB, error) {
	router := gin.Default()

	logger := log.GetLogger()

	db, err := postgres.New(cfg)
	if err != nil {
		logger.Warn("failed connect to db")
		return nil, nil, err
	}

	handler := handlers.NewHandler(cfg, db)
	auth := router.Group("/api/auth")
	{
		auth.POST("/login", handler.SignIn)    // вход
		auth.POST("/register", handler.SignUp) // регистрация
		auth.POST("/logout", handler.Logout)
		auth.POST("/refresh", handler.Refresh)
	}

	return router, db.DB, nil
}
