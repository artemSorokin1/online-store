package httpserver

import (
	"content_service/internal/api/product_client"
	"content_service/internal/api/seller_client"
	"content_service/internal/config"
	"net/http"

	"content_service/internal/handler"

	"github.com/go-chi/chi"
)

type HttpServer struct {
	router  chi.Router
	cfg     *config.ServerConfig
	handler *handler.Handler
}

func New(cfg *config.ServerConfig, sellerClient *seller_client.SellersClient, productClient *product_client.ProductsClient) *HttpServer {
	router := chi.NewRouter()

	return &HttpServer{
		router:  router,
		cfg:     cfg,
		handler: handler.New(sellerClient, productClient),
	}
}

func (s *HttpServer) setupRoutes() {
	s.router.Get("/api/content/search", s.handler.SearchHandler)
	s.router.Get("/api/content/products/{uuid}", s.handler.GetInfoHandler)
}

func (s *HttpServer) MustRun() {
	s.setupRoutes()

	if err := http.ListenAndServe(":8086", s.router); err != nil {
		panic("Failed to start HTTP server: " + err.Error())
	}
}
