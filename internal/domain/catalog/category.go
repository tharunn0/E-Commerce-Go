package catalog

import (
	"context"
	"time"
)

type CategoryRepository interface {
	CreateCategory(ctx context.Context, category *Category) (*Category, error)
	UpdateCategory(ctx context.Context, category *Category) (*Category, error)
	ChangeCategoryStatus(ctx context.Context, id int64, isActive bool) (*Category, error)
	GetCategoryByID(ctx context.Context, id int64, activeOnly bool) (*Category, error)
	GetAllCategories(ctx context.Context, activeOnly bool) ([]*Category, error)
	DeleteCategory(ctx context.Context, id int64) error
}

type Category struct {
	ID               int64     `json:"id"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	ParentCategoryID *int64    `json:"parent_category_id,omitempty"`
	IsActive         bool      `json:"is_active"`
	ImageURL         *string   `json:"image_url,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
type UpdateCategoryRequest struct {
	ID               int64   `json:"id"`
	Name             *string `json:"name"`
	Description      *string `json:"description"`
	ParentCategoryID *int64  `json:"parent_category_id,omitempty"`
	ImageURL         *string `json:"image_url,omitempty"`
}
type CategoryStatusRequest struct {
	IsActive bool `json:"is_active"`
}
