package catalog

import (
	"time"
)

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
