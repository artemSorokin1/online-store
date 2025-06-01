package service

import (
	"comment_service/internal/domain"
	"comment_service/internal/repository"
	"errors"
)

type RatingService struct {
	repo *repository.RatingRepository
}

func NewRatingService(repo *repository.RatingRepository) *RatingService {
	return &RatingService{repo: repo}
}

func (s *RatingService) CreateRating(rating *domain.Rating) error {
	if rating.Value < 1 || rating.Value > 5 {
		return errors.New("rating value must be between 1 and 5")
	}
	return s.repo.CreateRating(rating)
}

func (s *RatingService) GetProductRatings(productID uint) ([]domain.Rating, error) {
	return s.repo.GetProductRatings(productID)
}

func (s *RatingService) GetAverageRating(productID uint) (float64, error) {
	return s.repo.GetAverageRating(productID)
}

func (s *RatingService) DeleteRating(id uint) error {
	return s.repo.DeleteRating(id)
}

func (s *RatingService) UpdateRating(rating *domain.Rating) error {
	if rating.Value < 1 || rating.Value > 5 {
		return errors.New("rating value must be between 1 and 5")
	}
	return s.repo.UpdateRating(rating)
} 