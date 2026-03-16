package cart

import (
	"context"
	"errors"

	"github.com/tharunn0/E-Commerce-Go/internal/domain/promotion"
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
	Items          []*CartItem           `json:"items"`
	CartTotalPrice float64               `json:"cart_total_price"`
	CouponData     *promotion.CouponData `json:"coupon_data,omitempty"`
}
type CartItem struct {
	ProductVariantID int64    `json:"product_variant_id"`
	ProductID        int64    `json:"product_id"`
	CategoryID       int64    `json:"category_id"`
	ProductName      string   `json:"product_name"`
	SKU              string   `json:"sku"`
	OriginalPrice    float64  `json:"original_price"`
	SalePrice        *float64 `json:"sale_price"`
	Stock            int      `json:"stock"`
	ImageURL         string   `json:"image_url"`
	Quantity         int64    `json:"quantity"`
	TotalPrice       float64  `json:"total_price"`

	AppliedOffer *promotion.AppliedOfferData `json:"applied_offer,omitempty"`

	Status  CartItemStatus `json:"status,omitempty"`
	Message string         `json:"message,omitempty"`
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

func ApplyCouponToCart(cart *Cart, coupon *promotion.CouponResponse) error {
	if cart == nil || coupon == nil {
		return errors.New("invalid input")
	}

	var discount float64

	// calculate discount
	switch coupon.DiscountType {
	case "percentage":
		discount = (cart.CartTotalPrice * coupon.DiscountValue) / 100

	case "fixed":
		discount = coupon.DiscountValue

	default:
		return errors.New("invalid discount type")
	}

	couponData := &promotion.CouponData{}

	// max discount cap
	if coupon.MaxDiscountAmount > 0 && discount > coupon.MaxDiscountAmount {
		discount = coupon.MaxDiscountAmount
		couponData.Message = "Maximum discount applied"
	} else {
		couponData.Message = "Discount applied successfully"
	}

	if discount > cart.CartTotalPrice {
		discount = cart.CartTotalPrice
		couponData.Message = "Discount is greater than cart total"
	}

	cart.CartTotalPrice -= discount

	cart.CouponData = &promotion.CouponData{
		CouponCode:       coupon.CouponCode,
		DiscountType:     coupon.DiscountType,
		DiscountValue:    coupon.DiscountValue,
		DiscountedAmount: discount,
		Message:          couponData.Message,
	}

	return nil
}
