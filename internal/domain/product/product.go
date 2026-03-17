package product

import (
	"strings"
	"time"

	"github.com/tharunn0/E-Commerce-Go/internal/domain/review"
)

// brand

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

// product

type Product struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name" binding:"required"`
	BrandID     *int64    `json:"brand_id,omitempty"`
	Description string    `json:"description" binding:"required"`
	CategoryId  *int64    `json:"category_id" binding:"required"`
	MinPrice    float64   `json:"min_price" binding:"required"`
	MaxPrice    float64   `json:"max_price" binding:"required"`
	IsDigital   bool      `json:"is_digital"`
	IsActive    bool      `json:"is_active"`
	ImageURL    string    `json:"image_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateProductRequest struct {
	Name        string   `json:"name" binding:"required"`
	BrandID     *int64   `json:"brand_id,omitempty"`
	Description string   `json:"description" binding:"required"`
	CategoryId  *int64   `json:"category_id" binding:"required"`
	MinPrice    *float64 `json:"min_price,omitempty"`
	MaxPrice    *float64 `json:"max_price,omitempty"`
	IsDigital   bool     `json:"is_digital"`
	IsActive    bool     `json:"is_active"`
	ImageURL    string   `json:"image_url"`
}

type ProductStatusRequest struct {
	ID     int64
	Status bool `json:"is_active"`
}

type UpdateProductRequest struct {
	ID          int64    `json:"id"`
	Name        *string  `json:"name"`
	BrandID     *int64   `json:"brand_id,omitempty"`
	Description *string  `json:"description,omitempty"`
	CategoryId  *int64   `json:"category_id,omitempty"`
	MinPrice    *float64 `json:"min_price,omitempty"`
	MaxPrice    *float64 `json:"max_price,omitempty"`
	IsDigital   *bool    `json:"is_digital,omitempty"`
	IsActive    *bool    `json:"is_active,omitempty"`
	ImageURL    *string  `json:"image_url,omitempty"`
}

type ProductResponse struct {
	ID          int64                    `json:"id"`
	Name        string                   `json:"name"`
	Brand       *ProductBrandResponse    `json:"brand,omitempty"`
	Category    *ProductCategoryResponse `json:"category,omitempty"`
	Description string                   `json:"description"`
	MinPrice    float64                  `json:"min_price"`
	MaxPrice    float64                  `json:"max_price"`
	IsDigital   bool                     `json:"is_digital"`
	IsActive    *bool                    `json:"is_active,omitempty"`
	ImageURL    string                   `json:"image_url"`
	Rating      *float64                 `json:"rating,omitempty"`
	Reviews     []*review.Review         `json:"review,omitempty"`
	CreatedAt   *time.Time               `json:"created_at,omitempty"`
	UpdatedAt   *time.Time               `json:"updated_at,omitempty"`
}

type ProductCategoryResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type ProductBrandResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type ProductFilter struct {
	BrandID    *int    `form:"brand_id"`
	CategoryID *int    `form:"category_id"`
	Page       int     `form:"page"`
	Limit      int     `form:"limit"`
	Search     string  `form:"search"`
	Sort       string  `form:"sort"`
	Order      string  `form:"order"`
	MinPrice   float64 `form:"min_price"`
	MaxPrice   float64 `form:"max_price"`
}

func (f *ProductFilter) IsFiltersValid() bool {

	if f.Sort == "price" {
		f.Sort = "min_price"
	}

	if f.Sort == "" || f.Order == "" {
		return true
	}

	f.Order = strings.ToUpper(f.Order)
	f.Sort = strings.ToLower(f.Sort)

	validOrder := map[string]struct{}{
		"ASC":  {},
		"DESC": {},
	}
	validSortCol := map[string]struct{}{
		"min_price":  {},
		"name":       {},
		"created_at": {},
		"rating":     {},
	}

	if _, ok := validOrder[f.Order]; !ok {
		return false
	}
	if _, ok := validSortCol[f.Sort]; !ok {
		return false
	}

	return true
}
