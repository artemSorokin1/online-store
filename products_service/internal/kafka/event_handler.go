package kafka

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/artemSorokin1/products-grpc-api/internal/repository"
	"github.com/google/uuid"
)

type ProductEventHandler struct {
	repo *repository.ProductRepository
}

func NewProductEventHandler(repo *repository.ProductRepository) *ProductEventHandler {
	return &ProductEventHandler{
		repo: repo,
	}
}

func (h *ProductEventHandler) HandleProductEvent(ctx context.Context, event *ProductEvent) error {
	productID, err := strconv.ParseInt(event.ProductID, 10, 64)
	if err != nil {
		return fmt.Errorf("ошибка парсинга ID продукта: %v", err)
	}

	sellerID, err := uuid.Parse(event.Seller)
	if err != nil {
		return fmt.Errorf("ошибка парсинга ID продавца: %v", err)
	}

	info, err := json.Marshal(map[string]interface{}{
		"tags": event.Tags,
	})
	if err != nil {
		return fmt.Errorf("ошибка сериализации тегов: %v", err)
	}

	product := &repository.Product{
		ID:          productID,
		Name:        event.Name,
		Description: sql.NullString{String: event.Description, Valid: true},
		Info:        info,
		SellerID:    sellerID,
		Tags:        event.Tags,
	}

	switch event.EventType {
	case "created":
		return h.repo.CreateProduct(ctx, product)
	case "updated":
		return h.repo.UpdateProduct(ctx, product)
	case "deleted":
		return h.repo.DeleteProduct(ctx, productID)
	default:
		return fmt.Errorf("неизвестный тип события: %s", event.EventType)
	}
} 