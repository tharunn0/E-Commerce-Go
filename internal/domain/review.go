package domain

import (
	"context"
	"time"
)

type ReviewRepository interface {
	CreateReview(ctx context.Context, review *Review) (*Review, error)
	DeleteReview(ctx context.Context, id int64) error
}

type Review struct {
	ID          int64      `json:"id"`
	User        string     `json:"user"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Rating      float64    `json:"rating"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}
