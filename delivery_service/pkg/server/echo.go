package server

import (
	"context"
	"dlivery_service/delivery_service/internal/config"
	"dlivery_service/delivery_service/internal/repository/storage"
	"dlivery_service/delivery_service/internal/service/handlers"
	"dlivery_service/delivery_service/pkg/logger"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
)

type EchoServer struct {
	log     *logger.Logger
	server  *echo.Echo
	handler *handlers.Handler
}

func New(ctx context.Context, db *storage.DB) *EchoServer {
	logg := logger.GetLoggerFromContext(ctx)
	return &EchoServer{
		log:     logg,
		server:  echo.New(),
		handler: handlers.New(db),
	}
}

func (e *EchoServer) MustRun(cfg *config.Config) {
	e.server.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}))
	e.setHandlers()
	err := e.server.Start(fmt.Sprintf("0.0.0.0:%s", cfg.ServerCfg.Port))
	if err != nil {
		panic(err)
	}
}

func (e *EchoServer) setHandlers() {
	e.server.POST("/register", e.handler.RegisterUserHandler)
	e.server.POST("/login", e.handler.LoginUserHandler)
	e.server.POST("/logout", e.handler.LogoutUserHandler)
	e.server.GET("/products", e.handler.GetProductsHandler)
	e.server.GET("/products/:id", e.handler.GetProductByIdHandler)
	e.server.GET("/cart", e.handler.GetCartHandler)
	e.server.POST("/cart/add", e.handler.AddProductInCartHandler)
	e.server.POST("/cart/delete-item", e.handler.DeleteCartItemHandler)
	e.server.POST("/cart/clear", e.handler.DeleteCartHandler)
	e.server.POST("/checkout", e.handler.CheckoutHandler)
	e.server.POST("/logout", e.handler.LogoutUserHandler)
	e.server.POST("/admin/products", e.handler.AdminMiddleware(e.handler.AdminAddProductHandler))
	e.server.DELETE("/admin/products/:id", e.handler.AdminMiddleware(e.handler.AdminDeleteProductHandler))
}
