package postgres

import (
	"auth/internal/models"

	"github.com/google/uuid"
)

func (s *Storage) GetSellerByID(sellerID uuid.UUID) (*models.GrpcSeller, error) {
	var seller models.GrpcSeller
	err := s.DB.Get(&seller, "SELECT id, email, fullname, created_at FROM users WHERE role='seller' and id = $1", sellerID)
	if err != nil {
		return nil, err
	}
	return &seller, nil
}
