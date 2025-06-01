package models

type Product struct {
	ProductID   string   `json:"product_id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Seller      string   `json:"seller"`
}
