package review

import (
	"context"
	"errors"
	"time"
)

type ReviewRepository interface {
	CreateReview(ctx context.Context, review *CreateReviewRequest) error
	GetProductReviews(ctx context.Context, productID int64, filter *ReviewFilter) (*ProductReviews, error)
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

type ProductReviews struct {
	ProductID     int64    `json:"product_id"`
	ProductName   string   `json:"product_name"`
	AverageRating float64  `json:"average_rating"`
	TotalReviews  int64    `json:"total_reviews"`
	Reviews       []Review `json:"reviews"`
}

type ReviewFilter struct {
	SortBy  string `form:"sort_by"`
	OrderBy string `form:"order_by"`
	Page    int64  `form:"page"`
	Limit   int64  `form:"limit"`
}

type CreateReviewRequest struct {
	UserID      int64   `json:"user_id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Rating      float64 `json:"rating"`
	ProductID   int64   `json:"product_id"`
}

func (r *CreateReviewRequest) Validate() error {
	if r.Rating < 1 || r.Rating > 5 {
		return errors.New("Invalid rating")
	}
	if len(r.Title) < 5 || len(r.Title) > 50 {
		return errors.New("Title must be between 5 and 50 characters")
	}
	if len(r.Description) < 5 || len(r.Description) > 200 {
		return errors.New("Description must be between 5 and 200 characters")
	}
	if r.ProductID <= 0 {
		return errors.New("Invalid product ID")
	}
	return nil
}

func (f *ReviewFilter) Validate() error {
	if f.SortBy != "" {
		if f.SortBy != "rating" && f.SortBy != "created_at" {
			return errors.New("Invalid sort by")
		}
		f.SortBy = "created_at"
	}
	if f.OrderBy != "" {
		if f.OrderBy != "asc" && f.OrderBy != "desc" {
			return errors.New("Invalid order by")
		}
		f.OrderBy = "desc"
	}
	if f.Page <= 0 {
		f.Page = 1
	}
	if f.Limit <= 0 {
		f.Limit = 10
	}
	return nil
}
