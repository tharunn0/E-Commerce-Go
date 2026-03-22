package product

import (
	"context"
)

//go:generate mockgen -destination=mocks/mock_product_repo.go -package=mocks . ProductRepository
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
	ToggleProductStatus(ctx context.Context, req *ProductStatusRequest) error
	DeleteProduct(ctx context.Context, id int64) error

	// Variant operations
	//------------------
	CreateProductVariant(ctx context.Context, productVariant *CreateProductVariantRequest) (*ProductVariantResponse, error)
	GetVariantsByProductID(ctx context.Context, productID int64, activeOnly bool) (*ProductVariantBaseResponse, error)
	GetProductVariantByID(ctx context.Context, id int64, activeOnly bool) (*ProductVariantResponse, error)
	UpdateProductVariant(ctx context.Context, productVariant *UpdateProductVariantRequest) (*ProductVariant, error)
	ToggleVariantStatus(ctx context.Context, req *VariantStatusRequest) error
	DeleteProductVariant(ctx context.Context, id int64) error

	UpdateProductMinMaxPrice(ctx context.Context, productID int64) error

	GetProductVariantOrderInfo(ctx context.Context, productVariantID int64) (*VariantOrderInfo, error)

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
