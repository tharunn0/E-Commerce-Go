package domain

// import "context"

type OrderRepository interface {
	// ValidateCart(ctx context.Context, userID int64) (map[int64]int64, error)
}

type CartCheckoutRequest struct {
	AddressID    int64        `json:"address_id"`
	DeliveryType DeliveryType `json:"delivery_type"`
}
type CartCheckoutResponse struct {
	Cart                  *Cart        `json:"cart"`
	ShippingCost          float64      `json:"shipping_cost"`
	TotalAmount           float64      `json:"total_amount"`
	DeliveryType          DeliveryType `json:"delivery_type"`
	Address               *UserAddress `json:"address"`
	EstimatedDeliveryTime string       `json:"estimated_delivery_time"`
	EstimatedDeliveryDate string       `json:"estimated_delivery_date"`
}

type ProductVariantCheckoutRequest struct {
	ProductVariantID int64        `json:"product_variant_id"`
	Quantity         int64        `json:"quantity"`
	AddressID        int64        `json:"address_id"`
	DeliveryType     DeliveryType `json:"delivery_type"`
}
type ProductVariantCheckoutResponse struct {
	ProductVariant        *ProductVariantResponse `json:"product_variant"`
	ShippingCost          float64                 `json:"shipping_cost"`
	TotalAmount           float64                 `json:"total_amount"`
	DeliveryType          DeliveryType            `json:"delivery_type"`
	Address               *UserAddress            `json:"address"`
	EstimatedDeliveryTime string                  `json:"estimated_delivery_time"`
	EstimatedDeliveryDate string                  `json:"estimated_delivery_date"`
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
