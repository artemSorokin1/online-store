package dto

type ResponseDTO struct {
	ProductID        string `json:"id"`
	ProductName      string `json:"name"`
	ProductCreatedAt string `json:"created_at"`

	ProductDescription string   `json:"description,omitempty"`
	ProductPrice       int32    `json:"price,omitempty"`
	ProductImageURL    string   `json:"image_url,omitempty"`
	ProductInfo        string   `json:"info,omitempty"`
	ProductComments    []string `json:"comments,omitempty"`
	ProductRating      float64  `json:"rating,omitempty"`
	ProductTags        []string `json:"tags,omitempty"`

	SellerEmail    string `json:"email,omitempty"`
	SellerPhone    string `json:"phone,omitempty"`
	SellerFullName string `json:"full_name,omitempty"`
}
