package models

type Product struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int64  `json:"price"`
	ImageURL    string `json:"image_url"`
	SellerID    string `json:"seller_id"`
	Info        string `json:"info"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}
