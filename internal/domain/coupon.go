package domain

import (
	"context"
	"errors"
	"time"
)

type CouponRepository interface {
	CreateCoupon(ctx context.Context, req *CreateCouponRequest) (*CouponResponse, error)

	ListAllCoupons(ctx context.Context, filter *ListCouponsFilter) ([]CouponResponse, error)

	FetchCoupon(ctx context.Context, couponCode string) (*CouponResponse, error)
}

type CreateCouponRequest struct {
	CouponCode        string    `json:"coupon_code"`
	Description       string    `json:"description"`
	DiscountType      string    `json:"discount_type"`
	DiscountValue     float64   `json:"discount_value"`
	MinOrderAmount    float64   `json:"min_order_amount"`
	MaxDiscountAmount float64   `json:"max_discount_amount"`
	ValidFrom         time.Time `json:"valid_from"`
	ValidTo           time.Time `json:"valid_to"`
}

type CouponResponse struct {
	ID                int        `json:"id"`
	CouponCode        string     `json:"coupon_code"`
	Description       string     `json:"description"`
	DiscountType      string     `json:"discount_type"` // percentage or fixed
	DiscountValue     float64    `json:"discount_value"`
	MinOrderAmount    float64    `json:"min_order_amount"`
	MaxDiscountAmount float64    `json:"max_discount_amount"`
	ValidFrom         time.Time  `json:"valid_from"`
	ValidTo           time.Time  `json:"valid_to"`
	IsActive          bool       `json:"is_active"`
	CreatedAt         *time.Time `json:"created_at,omitempty"`
	UpdatedAt         *time.Time `json:"updated_at,omitempty"`
}

type ListCouponsFilter struct {
	CouponCode   string `form:"code"`
	DiscountType string `form:"type"`

	ValidFrom *time.Time `form:"valid_from" time_format:"2006-01-02"`
	ValidTo   *time.Time `form:"valid_to" time_format:"2006-01-02"`

	IsActive *bool `form:"active"`
}

type ApplyCouponRequest struct {
	CouponCode   string       `json:"coupon_code"`
	AddressID    int64        `json:"address_id"`
	DeliveryType DeliveryType `json:"delivery_type"`
}

type CouponData struct {
	CouponCode       string
	DiscountType     string
	DiscountValue    float64
	DiscountedAmount float64
	Message          string
}

func (req *CreateCouponRequest) Validate() error {
	if req.CouponCode == "" {
		return errors.New("coupon code is required")
	}
	if len(req.CouponCode) < 3 || len(req.CouponCode) > 10 {
		return errors.New("coupon code must be between 3 and 10 characters long")
	}
	if req.Description == "" {
		return errors.New("description is required")
	}
	if len(req.Description) < 10 {
		return errors.New("description must be at least 10 characters long")
	}
	if req.DiscountType != "percentage" && req.DiscountType != "fixed" {
		return errors.New("discount type is invalid")
	}
	if req.DiscountValue <= 0 {
		return errors.New("invalid discount value")
	}
	if req.MinOrderAmount <= 0 {
		return errors.New("invalid min order amount")
	}
	if req.MaxDiscountAmount <= 0 {
		return errors.New("invalid max discount amount")
	}
	if req.ValidFrom.IsZero() {
		return errors.New("invalid valid from")
	}
	if req.ValidTo.IsZero() {
		return errors.New("invalid valid to")
	}
	if req.ValidFrom.After(req.ValidTo) {
		return errors.New("valid from cannot be after valid to")
	}
	return nil
}

func ValidateCoupon(coupon *CouponResponse, now time.Time) error {
	if coupon == nil {
		return errors.New("coupon not found")
	}
	if coupon.IsActive == false {
		return errors.New("coupon is not active")
	}

	if now.Before(coupon.ValidFrom) {
		return errors.New("coupon is not yet valid")
	}
	if now.After(coupon.ValidTo) {
		return errors.New("coupon is expired")
	}
	return nil
}

func ApplyCouponToCart(cart *Cart, coupon *CouponResponse) error {
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

	couponData := &CouponData{}

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

	cart.CouponData = &CouponData{
		CouponCode:       coupon.CouponCode,
		DiscountType:     coupon.DiscountType,
		DiscountValue:    coupon.DiscountValue,
		DiscountedAmount: discount,
		Message:          couponData.Message,
	}

	return nil
}
