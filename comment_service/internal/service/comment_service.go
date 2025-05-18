package service

import (
	"comment_service/internal/domain"
	"comment_service/internal/repository"
)

type CommentService struct {
	repo *repository.CommentRepository
}

func NewCommentService(repo *repository.CommentRepository) *CommentService {
	return &CommentService{
		repo: repo,
	}
}

func (s *CommentService) CreateComment(comment *domain.Comment) error {
	return s.repo.Create(comment)
}

func (s *CommentService) GetProductComments(productID uint) ([]domain.Comment, error) {
	return s.repo.GetByProductID(productID)
}

func (s *CommentService) DeleteComment(id uint) error {
	return s.repo.Delete(id)
} 