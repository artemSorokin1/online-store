package service

import (
	"comment_service/internal/domain"
	"comment_service/internal/repository"
	"errors"
)

type CommentService struct {
	repo *repository.CommentRepository
}

func NewCommentService(repo *repository.CommentRepository) *CommentService {
	return &CommentService{repo: repo}
}

func (s *CommentService) CreateComment(comment *domain.Comment) error {
	if comment.Rating != 0 && (comment.Rating < 1 || comment.Rating > 5) {
		return errors.New("rating value must be between 1 and 5")
	}
	return s.repo.CreateComment(comment)
}

func (s *CommentService) GetProductComments(productID uint) ([]domain.Comment, error) {
	return s.repo.GetProductComments(productID)
}

func (s *CommentService) DeleteComment(id uint) error {
	return s.repo.DeleteComment(id)
}

func (s *CommentService) UpdateComment(comment *domain.Comment) error {
	if comment.Rating != 0 && (comment.Rating < 1 || comment.Rating > 5) {
		return errors.New("rating value must be between 1 and 5")
	}
	return s.repo.UpdateComment(comment)
}

func (s *CommentService) GetAverageRating(productID uint) (float64, error) {
	return s.repo.GetAverageRating(productID)
} 