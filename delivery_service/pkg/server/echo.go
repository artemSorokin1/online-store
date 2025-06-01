package server

import (
	"context"
	"dlivery_service/delivery_service/internal/config"
	"dlivery_service/delivery_service/internal/repository/storage"
	"dlivery_service/delivery_service/internal/service/handlers"
	"dlivery_service/delivery_service/pkg/metrics"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type EchoServer struct {
	logger  *zap.Logger
	server  *echo.Echo
	handler *handlers.Handler
}

func New(ctx context.Context, db *storage.DB, logger *zap.Logger) *EchoServer {
	return &EchoServer{
		logger:  logger,
		server:  echo.New(),
		handler: handlers.New(logger, db),
	}
}

func (e *EchoServer) MustRun(cfg *config.Config) {
	e.logger.Info("starting server")
	e.server.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	e.logger.Debug("setting handlers")
	e.setHandlers()
	e.logger.Debug("handlers set")

	err := e.server.Start(fmt.Sprintf("0.0.0.0:%s", cfg.ServerCfg.Port))
	if err != nil {
		e.logger.Fatal("failed to start server", zap.Error(err))
	}

	e.logger.Info("server started", zap.String("port", cfg.ServerCfg.Port))
}

func (e *EchoServer) setHandlers() {
	am := metrics.NewAuthMetrics()

	auth := e.server.Group("/api/auth", echo.WrapMiddleware(am.Middleware))
	{
		auth.POST("/register", e.handler.RegisterUserHandler)
		auth.POST("/login", e.handler.LoginUserHandler)
		auth.POST("/logout", e.handler.LogoutUserHandler)
	}

	pm := metrics.NewProductsMetrics()
	products := e.server.Group("/api/products", echo.WrapMiddleware(pm.Middleware))
	{
		products.GET("/", e.handler.GetProductsHandler)
		products.GET("/:id", e.handler.GetProductByIdHandler)
	}

	cm := metrics.NewCartMetrics()
	cart := e.server.Group("/api/cart", echo.WrapMiddleware(cm.Middleware))
	{
		cart.GET("/", e.handler.GetCartHandler)
		cart.POST("/add", e.handler.AddProductInCartHandler)
		cart.POST("/delete-item", e.handler.DeleteCartItemHandler)
		cart.POST("/clear", e.handler.DeleteCartHandler)
	}

	admin := e.server.Group("/api/admin", e.handler.AdminMiddleware)
	{
		admin.POST("/products", e.handler.AdminAddProductHandler)
		admin.DELETE("/products/:id", e.handler.AdminDeleteProductHandler)
	}

	e.server.POST("/checkout", e.handler.CheckoutHandler)

	e.server.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
}
