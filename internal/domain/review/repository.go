package review

import (
	"context"
)

type ReviewRepository interface {
	CreateReview(ctx context.Context, review *CreateReviewRequest) error
	GetProductReviews(ctx context.Context, productID int64, filter *ReviewFilter) (*ProductReviews, error)
	DeleteReview(ctx context.Context, id int64) error
}
