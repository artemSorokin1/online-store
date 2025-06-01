package repository

import (
	"comment_service/internal/domain"
	"gorm.io/gorm"
	"time"
)

type RatingRepository struct {
	db *gorm.DB
}

func NewRatingRepository(db *gorm.DB) *RatingRepository {
	return &RatingRepository{db: db}
}

func (r *RatingRepository) CreateRating(rating *domain.Rating) error {
	rating.CreatedAt = time.Now()
	rating.UpdatedAt = time.Now()
	return r.db.Create(rating).Error
}

func (r *RatingRepository) GetProductRatings(productID uint) ([]domain.Rating, error) {
	var ratings []domain.Rating
	err := r.db.Where("product_id = ?", productID).Find(&ratings).Error
	return ratings, err
}

func (r *RatingRepository) GetAverageRating(productID uint) (float64, error) {
	var avg float64
	err := r.db.Model(&domain.Rating{}).
		Where("product_id = ?", productID).
		Select("COALESCE(AVG(value), 0)").
		Scan(&avg).Error
	return avg, err
}

func (r *RatingRepository) DeleteRating(id uint) error {
	return r.db.Delete(&domain.Rating{}, id).Error
}

func (r *RatingRepository) UpdateRating(rating *domain.Rating) error {
	// Сначала получаем существующую оценку
	var existingRating domain.Rating
	if err := r.db.First(&existingRating, rating.ID).Error; err != nil {
		return err
	}

	// Обновляем только value и updated_at
	rating.ProductID = existingRating.ProductID
	rating.UserID = existingRating.UserID
	rating.CreatedAt = existingRating.CreatedAt
	rating.UpdatedAt = time.Now()

	// Обновляем запись
	if err := r.db.Model(&domain.Rating{}).Where("id = ?", rating.ID).Updates(map[string]interface{}{
		"value":      rating.Value,
		"updated_at": rating.UpdatedAt,
	}).Error; err != nil {
		return err
	}

	// Получаем обновленную запись
	return r.db.First(rating, rating.ID).Error
} 