package service

import (
	"log"

	"searchservice/internal/models"
	"searchservice/internal/repository"
)

type ProductService struct {
	repo *repository.ElasticRepository
}

func NewProductService(repo *repository.ElasticRepository) *ProductService {
	return &ProductService{repo: repo}
}

func (ps *ProductService) HandleNewProduct(prod *models.Product) {
	if prod.ProductID == "" {
		log.Printf("Skipping product with empty ID: %#v\n", prod)
		return
	}
	if err := ps.repo.IndexProduct(prod); err != nil {
		log.Printf("Failed to index product ID=%s: %s", prod.ProductID, err)
	}
}

func (ps *ProductService) Search(q string) ([]string, error) {
	return ps.repo.SearchProductIDs(q)
}
