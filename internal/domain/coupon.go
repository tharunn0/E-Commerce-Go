package domain

import (
	"context"
	"errors"
	"time"
)

type CouponRepository interface {
	CreateCoupon(ctx context.Context, req *CreateCouponRequest) (*CouponResponse, error)
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
	DiscountType      string     `json:"discount_type"`
	DiscountValue     float64    `json:"discount_value"`
	MinOrderAmount    float64    `json:"min_order_amount"`
	MaxDiscountAmount float64    `json:"max_discount_amount"`
	ValidFrom         time.Time  `json:"valid_from"`
	ValidTo           time.Time  `json:"valid_to"`
	IsActive          bool       `json:"is_active"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         *time.Time `json:"updated_at,omitempty"`
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
