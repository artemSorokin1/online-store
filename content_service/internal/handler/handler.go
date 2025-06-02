package handler

import (
	"content_service/dto"
	"content_service/internal/api/product_client"
	"content_service/internal/api/seller_client"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi"
)

type Handler struct {
	sellerClient  *seller_client.SellersClient
	productClient *product_client.ProductsClient
}

func New(sellerClient *seller_client.SellersClient, producClient *product_client.ProductsClient) *Handler {
	return &Handler{
		sellerClient:  sellerClient,
		productClient: producClient,
	}
}

func (h *Handler) SearchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if query == "" {
		http.Error(w, "missing query parameter", http.StatusBadRequest)
		return
	}

	searchURL := fmt.Sprintf("http://search_service:8085/search?q=%s", url.QueryEscape(query))

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		http.Error(w, "failed to create request to search service", http.StatusInternalServerError)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "error calling search service: "+err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf("search service returned status %d", resp.StatusCode), http.StatusBadGateway)
		return
	}

	var result struct {
		ProductIDs []string `json:"product_ids"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		http.Error(w, "failed to decode search response: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if len(result.ProductIDs) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("[]"))
		return
	}

	response, err := h.GetInfoFromServices(result.ProductIDs)
	if err != nil {
		http.Error(w, "failed to get product info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode response: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetInfoHandler(w http.ResponseWriter, r *http.Request) {
	uuidParam := chi.URLParam(r, "uuid")
	if uuidParam == "" {
		http.Error(w, "missing uuids parameter", http.StatusBadRequest)
		return
	}

	// Разбиваем строку на срез UUID
	uuids := make([]string, 1)
	uuids[0] = uuidParam

	response, err := h.GetInfoFromServices(uuids)
	if err != nil {
		http.Error(w, "failed to get product info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode response: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetInfoFromServices(uuids []string) ([]dto.ResponseDTO, error) {
	var response []dto.ResponseDTO
	for _, uuid := range uuids {
		product, err := h.productClient.GetProduct(context.Background(), uuid)
		if err != nil {
			return []dto.ResponseDTO{}, fmt.Errorf("failed to get product %s: %w", uuid, err)
		}

		seller, err := h.sellerClient.GetSeller(context.Background(), product.SellerId)
		if err != nil {
			return []dto.ResponseDTO{}, fmt.Errorf("failed to get seller for product %s: %w", uuid, err)
		}

		fmt.Println("seller:", seller)

		response = append(response, dto.ResponseDTO{
			ProductID:          product.Id,
			ProductName:        product.Name,
			ProductCreatedAt:   product.CreatedAt.String(),
			ProductDescription: product.Description,
			ProductPrice:       product.Price,
			ProductImageURL:    product.ImageUrl,
			ProductInfo:        product.Info.String(),
			ProductComments:    product.Comments,
			ProductRating:      product.Rating,
			ProductTags:        product.Tags,
			SellerEmail:        seller.Email,
			SellerPhone:        seller.Phone,
			SellerFullName:     seller.Fullname,
		})
	}

	fmt.Println(response)
	return response, nil
}
