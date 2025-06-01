package repository

import (
	"comment_service/internal/domain"
	"gorm.io/gorm"
	"time"
)

type CommentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) CreateComment(comment *domain.Comment) error {
	comment.CreatedAt = time.Now()
	comment.UpdatedAt = time.Now()
	return r.db.Create(comment).Error
}

func (r *CommentRepository) GetProductComments(productID uint) ([]domain.Comment, error) {
	var comments []domain.Comment
	err := r.db.Where("product_id = ?", productID).Find(&comments).Error
	return comments, err
}

func (r *CommentRepository) DeleteComment(id uint) error {
	return r.db.Delete(&domain.Comment{}, id).Error
}

func (r *CommentRepository) UpdateComment(comment *domain.Comment) error {
	// Сначала получаем существующий комментарий
	var existingComment domain.Comment
	if err := r.db.First(&existingComment, comment.ID).Error; err != nil {
		return err
	}

	// Обновляем только изменяемые поля
	comment.ProductID = existingComment.ProductID
	comment.UserID = existingComment.UserID
	comment.CreatedAt = existingComment.CreatedAt
	comment.UpdatedAt = time.Now()

	// Если рейтинг не указан, сохраняем старый
	if comment.Rating == 0 {
		comment.Rating = existingComment.Rating
	}

	// Если контент пустой, сохраняем старый
	if comment.Content == "" {
		comment.Content = existingComment.Content
	}

	// Обновляем запись
	if err := r.db.Model(&domain.Comment{}).Where("id = ?", comment.ID).Updates(map[string]interface{}{
		"content":    comment.Content,
		"rating":     comment.Rating,
		"updated_at": comment.UpdatedAt,
	}).Error; err != nil {
		return err
	}

	// Получаем обновленную запись
	return r.db.First(comment, comment.ID).Error
}

func (r *CommentRepository) GetAverageRating(productID uint) (float64, error) {
	var avg float64
	err := r.db.Model(&domain.Comment{}).
		Where("product_id = ? AND rating > 0", productID).
		Select("COALESCE(AVG(rating), 0)").
		Scan(&avg).Error
	return avg, err
} 