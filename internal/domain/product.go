package domain

import (
	"context"
	"time"
)

type ProductRepository interface {
	// Brand operations
	CreateBrand(ctx context.Context, createBrandRequest *CreateBrandRequest) (*Brand, error)
	GetAllBrands(ctx context.Context, activeOnly bool) ([]*Brand, error)
	GetBrandByID(ctx context.Context, id int64, activeOnly bool) (*Brand, error)
	UpdateBrand(ctx context.Context, brand *Brand) (*Brand, error)
	DeleteBrand(ctx context.Context, id int64) error
}

type Brand struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	LogoURL     *string   `json:"logo_url,omitempty"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateBrandRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	LogoURL     *string `json:"logo_url,omitempty"`
}

type UpdateBrandRequest struct {
	ID          int64   `json:"id"`
	Name        *string `json:"name"`
	Description *string `json:"description,omitempty"`
	LogoURL     *string `json:"logo_url,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
}
