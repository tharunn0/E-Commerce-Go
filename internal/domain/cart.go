package domain

import (
	"context"
)

type CartItemStatus string

const (
	StatusOutOfStock CartItemStatus = "OUT_OF_STOCK"
	StatusLowStock   CartItemStatus = "LOW_STOCK"
	StatusInStock    CartItemStatus = "IN_STOCK"
)

type CartRepository interface {
	AddToCart(ctx context.Context, userID int64, productVariantID int64, quantity int64) (*int64, error)
	GetCartByID(ctx context.Context, cartID int64) (*Cart, error)
	GetCartByUserID(ctx context.Context, userID int64) (*Cart, error)
	RemoveCartItem(ctx context.Context, userID int64, productVariantID int64) (int64, error)
	UpdateCartItemQuantity(ctx context.Context, userID int64, req *UpdateCartItemQuantityRequest) (int64, error)
	EmptyCart(ctx context.Context, userID int64) error
	GetCartVariantInfo(ctx context.Context, userID int64) (map[int64]VariantInfo, error)
}
type VariantInfo struct {
	ProductName string `json:"product_name"`
	SKU         string `json:"sku"`
	Stock       int64  `json:"stock"`
}

type Cart struct {
	Items          []*CartItem `json:"items"`
	CartTotalPrice float64     `json:"cart_total_price"`
}

type CartItem struct {
	ProductVariantID int64          `json:"product_variant_id"`
	ProductID        int64          `json:"product_id"`
	CategoryID       int64          `json:"category_id"`
	ProductName      string         `json:"product_name"`
	SKU              string         `json:"sku"`
	OriginalPrice    float64        `json:"original_price"`
	SalePrice        *float64       `json:"sale_price"`
	Stock            int            `json:"stock"`
	ImageURL         string         `json:"image_url"`
	Quantity         int64          `json:"quantity"`
	TotalPrice       float64        `json:"total_price"`
	Status           CartItemStatus `json:"status,omitempty"`
	Message          string         `json:"message,omitempty"`
}

type AddToCartRequest struct {
	ProductVariantID int64 `json:"product_variant_id"`
	Quantity         int64 `json:"quantity"`
}

type UpdateCartItemQuantityRequest struct {
	ProductVariantID int64 `json:"product_variant_id"`
	Quantity         int64 `json:"quantity"`
}
type RemoveCartItemRequest struct {
	ProductVariantID int64 `json:"product_variant_id"`
}
