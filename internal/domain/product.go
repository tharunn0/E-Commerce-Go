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

	// Product operations
	CreateProduct(ctx context.Context, product *CreateProductRequest) (*Product, error)
	GetProducts(ctx context.Context, filter *ProductFilter, activeOnly bool) ([]*ProductResponse, int64, error)
	GetProductByID(ctx context.Context, id int64, activeOnly bool) (*ProductResponse, error)
	UpdateProduct(ctx context.Context, updateProductRequest *UpdateProductRequest) error
	DeleteProduct(ctx context.Context, id int64) error

	// Variant operations
	//------------------
	CreateProductVariant(ctx context.Context, productVariant *CreateProductVariantRequest) (*ProductVariantResponse, error)
	// GetProductVariants(ctx context.Context, productID int64) ([]ProductVariant, error)
	GetProductVariantByID(ctx context.Context, id int64, activeOnly bool) (*ProductVariantResponse, error)
	UpdateProductVariant(ctx context.Context, productVariant *UpdateProductVariantRequest) (*ProductVariant, error)
	DeleteProductVariant(ctx context.Context, id int64) error

	// Attribute operations
	//------------------
	CreateAttribute(ctx context.Context, attribute *CreateAttributeRequest) (*Attribute, error)
	AddAttributeValues(ctx context.Context, attributeValues *AddAttributeValuesRequest) error
	GetAttributes(ctx context.Context, activeOnly bool) ([]*Attribute, error)
	GetAttributeByID(ctx context.Context, id int64, activeOnly bool) (*Attribute, error)
	// UpdateAttribute(ctx context.Context, attribute *Attribute) (*Attribute, error)
	DeleteAttribute(ctx context.Context, id int64) error
	DeleteAttributeValues(ctx context.Context, req *DeleteAttributeValuesRequest) error
}

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
	BasePrice   float64   `json:"base_price" binding:"required"`
	IsDigital   bool      `json:"is_digital"`
	IsActive    bool      `json:"is_active"`
	ImageURL    string    `json:"image_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateProductRequest struct {
	Name        string  `json:"name" binding:"required"`
	BrandID     *int64  `json:"brand_id,omitempty"`
	Description string  `json:"description" binding:"required"`
	CategoryId  *int64  `json:"category_id" binding:"required"`
	BasePrice   float64 `json:"base_price" binding:"required"`
	IsDigital   bool    `json:"is_digital"`
	IsActive    bool    `json:"is_active"`
	ImageURL    string  `json:"image_url"`
}

type UpdateProductRequest struct {
	ID          int64    `json:"id"`
	Name        *string  `json:"name"`
	BrandID     *int64   `json:"brand_id,omitempty"`
	Description *string  `json:"description,omitempty"`
	CategoryId  *int64   `json:"category_id,omitempty"`
	BasePrice   *float64 `json:"base_price,omitempty"`
	IsDigital   *bool    `json:"is_digital,omitempty"`
	IsActive    *bool    `json:"is_active,omitempty"`
	ImageURL    *string  `json:"image_url,omitempty"`
}

type ProductResponse struct {
	ID          int64                   `json:"id"`
	Name        string                  `json:"name"`
	Brand       ProductBrandResponse    `json:"brand"`
	Category    ProductCategoryResponse `json:"category"`
	Description string                  `json:"description"`
	BasePrice   float64                 `json:"base_price"`
	IsDigital   bool                    `json:"is_digital"`
	IsActive    bool                    `json:"is_active"`
	ImageURL    string                  `json:"image_url"`
	CreatedAt   time.Time               `json:"created_at"`
	UpdatedAt   time.Time               `json:"updated_at"`
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
	Page   int    `json:"page"`
	Limit  int    `json:"limit"`
	Search string `json:"search"`
	Sort   string `json:"sort"`
	Order  string `json:"order"`
}
