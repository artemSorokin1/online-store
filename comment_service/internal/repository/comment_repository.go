package repository

import (
	"comment_service/internal/domain"
	"gorm.io/gorm"
)

type CommentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) Create(comment *domain.Comment) error {
	return r.db.Create(comment).Error
}

func (r *CommentRepository) GetByProductID(productID uint) ([]domain.Comment, error) {
	var comments []domain.Comment
	err := r.db.Where("product_id = ?", productID).Order("created_at desc").Find(&comments).Error
	return comments, err
}

func (r *CommentRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Comment{}, id).Error
} 