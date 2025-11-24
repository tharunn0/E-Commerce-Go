package domain

import "context"

type OrderRepository interface {
	ValidateCart(ctx context.Context, userID int64) (map[int64]int64, error)
}

type ProductVariantCheckoutRequest struct {
	ProductVariantID int64 `json:"product_variant_id"`
	Quantity         int64 `json:"quantity"`
}

type VariantStock struct {
	ProductVariantID int64 `json:"product_variant_id"`
	Stock            int64 `json:"stock"`
}

type NotEnoughStockError struct {
	ProductVariantID int64  `json:"product_variant_id"`
	SKU              string `json:"sku"`
	Quantity         int64  `json:"quantity"`
	Stock            int64  `json:"available_stock"`
}
