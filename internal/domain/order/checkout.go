package order

import (
	"github.com/tharunn0/E-Commerce-Go/internal/domain/cart"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/product"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/shipping"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/user"
)

type CartCheckoutRequest struct {
	AddressID    int64                 `json:"address_id"`
	DeliveryType shipping.DeliveryType `json:"delivery_type"`
}
type CartCheckoutResponse struct {
	Cart                  *cart.Cart            `json:"cart"`
	ShippingCost          float64               `json:"shipping_cost"`
	TotalAmount           float64               `json:"total_amount"`
	DeliveryType          shipping.DeliveryType `json:"delivery_type"`
	Address               *user.UserAddress     `json:"address"`
	EstimatedDeliveryTime string                `json:"estimated_delivery_time"`
	EstimatedDeliveryDate string                `json:"estimated_delivery_date"`
}

type ProductVariantCheckoutRequest struct {
	ProductVariantID int64 `json:"product_variant_id"`

	Quantity     int64                 `json:"quantity"`
	AddressID    int64                 `json:"address_id"`
	DeliveryType shipping.DeliveryType `json:"delivery_type"`
}
type ProductVariantCheckoutResponse struct {
	ProductVariant        *product.ProductVariantResponse `json:"product_variant"`
	ShippingCost          float64                         `json:"shipping_cost"`
	TotalAmount           float64                         `json:"total_amount"`
	DeliveryType          shipping.DeliveryType           `json:"delivery_type"`
	Address               *user.UserAddress               `json:"address"`
	EstimatedDeliveryTime string                          `json:"estimated_delivery_time"`
	EstimatedDeliveryDate string                          `json:"estimated_delivery_date"`
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
