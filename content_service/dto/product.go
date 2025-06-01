package dto

type ProductDTO struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Price       int64    `json:"price"`
	ImageURL    string   `json:"image_url"`
	Info        string   `json:"info"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
	SellerID    string   `json:"seller_id"`
	Comments    []string `json:"comments"`
	Rating      float64  `json:"rating"`
	Tags        []string `json:"tags"`
}
