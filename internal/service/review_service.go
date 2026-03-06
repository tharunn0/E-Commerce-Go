package service

import (
	"context"

	"github.com/tharunn0/E-Commerce-Go/internal/domain"
	"go.uber.org/zap"
)

type ReviewService struct {
	repo domain.ReviewRepository
	log  *zap.Logger
}

func NewReviewService(repo domain.ReviewRepository, log *zap.Logger) *ReviewService {
	return &ReviewService{repo: repo, log: log}
}

func (s *ReviewService) CreateReview(ctx context.Context, review *domain.Review) (*domain.Review, error) {
	return s.repo.CreateReview(ctx, review)
}

func (s *ReviewService) DeleteReview(ctx context.Context, id int64) error {
	return s.repo.DeleteReview(ctx, id)
}
