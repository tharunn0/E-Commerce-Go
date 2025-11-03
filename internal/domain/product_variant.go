package domain

import "time"

type ProductVariant struct {
	ID              int64     `json:"id"`
	ProductID       int64     `json:"product_id"`
	SKU             string    `json:"sku"`
	PriceDifference float64   `json:"price_difference"`
	Stock           int       `json:"stock"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`

	VariantImages     []ProductVariantImage `json:"variant_images"`
	VariantAttributes []VariantAttribute    `json:"variant_attributes,omitempty"`
}
type ProductVariantImage struct {
	ID               int64     `json:"id"`
	ProductVariantID int64     `json:"product_variant_id"`
	URL              string    `json:"url"`
	CreatedAt        time.Time `json:"created_at"`
}
type ProductVariantAttributeRequest struct {
	AttributeID      *int64 `json:"attribute_id"`
	AttributeValueID *int64 `json:"attribute_value_id"`
}
type CreateProductVariantRequest struct {
	ProductID         int64                            `json:"product_id"`
	SKU               string                           `json:"sku"`
	PriceDifference   float64                          `json:"price_difference"`
	Stock             int                              `json:"stock"`
	Images            []string                         `json:"images"`
	VariantAttributes []ProductVariantAttributeRequest `json:"attributes"`
}

type VariantAttribute struct {
	ID               int64     `json:"id"`
	ProductVariantID int64     `json:"product_variant_id"`
	AttributeID      int64     `json:"attribute_id"`
	AttributeValueID int64     `json:"attribute_value_id"`
	CreatedAt        time.Time `json:"created_at"`
}
type UpdateProductVariantRequest struct {
	ID              int64    `json:"id"`
	ProductID       *int64   `json:"product_id,omitempty"`
	SKU             *string  `json:"sku,omitempty"`
	PriceDifference *float64 `json:"price_difference,omitempty"`
	Stock           *int     `json:"stock,omitempty"`
	IsActive        *bool    `json:"is_active,omitempty"`
}

type ProductVariantResponse struct {
	ID              int64                    `json:"id"`
	BaseProduct     BaseProduct              `json:"base_product"`
	SKU             string                   `json:"sku"`
	PriceDifference float64                  `json:"price_difference"`
	TotalPrice      float64                  `json:"total_price"`
	Stock           int                      `json:"stock"`
	VariantImages   []string                 `json:"variant_images"`
	Attributes      []AttributeValueResponse `json:"attributes"`
	CreatedAt       time.Time                `json:"created_at"`
}

type AttributeValueResponse struct {
	ID        int64  `json:"id"`
	Attribute string `json:"attribute"`
	Value     string `json:"value"`
}

type BaseProduct struct {
	ID        int64                `json:"id"`
	Name      string               `json:"name"`
	Brand     ProductBrandResponse `json:"brand"`
	BasePrice float64              `json:"base_price"`
}

// Attribute models
// //////////////////////////////////////////////////////////
type Attribute struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	DataType  string    `json:"data_type"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`

	Values []AttributeValue `json:"values,omitempty"`
}
type AttributeValue struct {
	ID          int64     `json:"id"`
	AttributeID *int64    `json:"attribute_id,omitempty"`
	Value       *string   `json:"value,omitempty"`
	IsActive    *bool     `json:"is_active,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type CreateAttributeRequest struct {
	Name     string           `json:"name"`
	DataType string           `json:"data_type"`
	IsActive *bool            `json:"is_active,omitempty"`
	Values   []AttributeValue `json:"values"`
}
type AddAttributeValuesRequest struct {
	AttributeID int64    `json:"attribute_id"`
	Values      []string `json:"values"`
}

type DeleteAttributeValuesRequest struct {
	ValueIDs []int64 `json:"value_ids"`
}
