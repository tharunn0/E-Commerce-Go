package domain

import "time"

type ProductVariant struct { // update product variant
	ID            int64     `json:"id"`
	ProductID     int64     `json:"product_id"`
	SKU           string    `json:"sku"`
	OriginalPrice float64   `json:"original_price"`
	SalePrice     *float64  `json:"sale_price"`
	Stock         int       `json:"stock"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`

	VariantImages     []ProductVariantImage `json:"variant_images"`
	VariantAttributes []VariantAttribute    `json:"variant_attributes,omitempty"`
}

type ProductVariantResponse struct { // get variant by id, create variant resposne
	ID            int64                    `json:"id"`
	BaseProduct   *BaseProduct             `json:"base_product,omitempty"`
	SKU           string                   `json:"sku"`
	OriginalPrice float64                  `json:"original_price"`
	SalePrice     *float64                 `json:"sale_price"`
	Stock         int                      `json:"stock"`
	VariantImages []string                 `json:"variant_images"`
	Attributes    []AttributeValueResponse `json:"attributes"`
	CreatedAt     time.Time                `json:"created_at"`
}
type VariantBaseResponse struct { // get variants by product id
	ID            int64                    `json:"id"`
	SKU           string                   `json:"sku"`
	OriginalPrice float64                  `json:"original_price"`
	SalePrice     *float64                 `json:"sale_price"`
	Stock         int                      `json:"stock"`
	IsActive      bool                     `json:"is_active"`
	Images        []string                 `json:"images"`
	Attributes    []AttributeValueResponse `json:"attributes"`
	CreatedAt     time.Time                `json:"created_at"`
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
	OriginalPrice     float64                          `json:"original_price"`
	SalePrice         *float64                         `json:"sale_price"`
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
	ID            int64    `json:"id"`
	ProductID     *int64   `json:"product_id,omitempty"`
	SKU           *string  `json:"sku,omitempty"`
	OriginalPrice *float64 `json:"original_price,omitempty"`
	SalePrice     *float64 `json:"sale_price"`
	Stock         *int     `json:"stock,omitempty"`
	IsActive      *bool    `json:"is_active,omitempty"`
}

type VariantStatusRequest struct {
	ID     int64
	Status bool `json:"is_active"`
}

type AttributeValueResponse struct {
	ID        int64  `json:"id"`
	Attribute string `json:"attribute"`
	Value     string `json:"value"`
}

type BaseProduct struct {
	ID    int64                `json:"id"`
	Name  string               `json:"name"`
	Brand ProductBrandResponse `json:"brand"`
}

type ProductVariantBaseResponse struct {
	ProductID   int64                 `json:"product_id"`
	ProductName string                `json:"product_name"`
	BrandName   string                `json:"brand_name"`
	MinPrice    float64               `json:"min_price"`
	MaxPrice    float64               `json:"max_price"`
	IsDigital   bool                  `json:"is_digital"`
	Variants    []VariantBaseResponse `json:"variants"`
	ImageURL    string                `json:"image_url"`
	CreatedAt   time.Time             `json:"created_at"`
}

type VariantOrderInfo struct {
	ProductName      string `json:"product_name"`
	ProductVariantID int64  `json:"product_variant_id"`
	SKU              string `json:"sku"`
}

// Attribute models
// //////////////////////////////////////////////////////////
type Attribute struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	DataType string `json:"data_type"`
	IsActive bool   `json:"is_active"`

	Values []AttributeValue `json:"values,omitempty"`
}
type AttributeValue struct {
	ID          int64   `json:"id"`
	AttributeID *int64  `json:"attribute_id,omitempty"`
	Value       *string `json:"value,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
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

type DeleteAttributeRequest struct {
	ID int64
}
type DeleteAttributeValuesRequest struct {
	ValueIDs []int64 `json:"value_ids"`
}
