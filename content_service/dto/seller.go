package dto

type SellerDTO struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	FullName  string `json:"full_name"`
	CreatedAt string `json:"created_at"`
}
