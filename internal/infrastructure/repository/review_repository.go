package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tharunn0/E-Commerce-Go/internal/domain"
)

type ReviewRepository struct {
	DB *pgxpool.Pool
}

func NewReviewRepository(db *pgxpool.Pool) *ReviewRepository {
	return &ReviewRepository{DB: db}
}

func (r *ReviewRepository) CreateReview(ctx context.Context, review *domain.Review) (*domain.Review, error) {
	return nil, nil
}

func (r *ReviewRepository) DeleteReview(ctx context.Context, id int64) error {
	return nil
}
