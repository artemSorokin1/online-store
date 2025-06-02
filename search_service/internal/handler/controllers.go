package handler

import (
	"encoding/json"
	"net/http"
	"searchservice/internal/service"
)

type SearchHandler struct {
	service *service.ProductService
}

func NewSearchHandler(svc *service.ProductService) *SearchHandler {
	return &SearchHandler{service: svc}
}

// ServeHTTP обрабатывает GET /search?q=…
func (h *SearchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET is allowed", http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query().Get("q")
	if q == "" {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	resultIDs, err := h.service.Search(q)
	if err != nil {
		http.Error(w, "Internal server error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"product_ids": resultIDs})
}
