package promotion

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/tharunn0/E-Commerce-Go/internal/domain/shipping"
)

type CouponRepository interface {
	CreateCoupon(ctx context.Context, req *CreateCouponRequest) (*CouponResponse, error)

	ListAllCoupons(ctx context.Context, filter *ListCouponsFilter) ([]CouponResponse, error)

	ListReferralRewards(ctx context.Context, userID int64) ([]CouponResponse, error)

	FetchCoupon(ctx context.Context, couponCode string, userID int64) (*CouponResponse, error)
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
	CouponCode   string                `json:"coupon_code"`
	AddressID    int64                 `json:"address_id"`
	DeliveryType shipping.DeliveryType `json:"delivery_type"`
}

type CouponData struct {
	CouponCode       string
	DiscountType     string
	DiscountValue    float64
	DiscountedAmount float64
	Message          string
}

func (req *CreateCouponRequest) Validate() error {
	code := strings.ToUpper(strings.TrimSpace(req.CouponCode))

	// --- Coupon code ---
	if code == "" {
		return errors.New("coupon code is required")
	}
	if len(code) < 3 || len(code) > 10 {
		return errors.New("coupon code must be between 3 and 10 characters long")
	}
	if !regexp.MustCompile(`^[A-Z0-9]+$`).MatchString(code) {
		return errors.New("coupon code must be uppercase alphanumeric")
	}
	req.CouponCode = code

	if strings.TrimSpace(req.Description) == "" {
		return errors.New("description is required")
	}
	if len(req.Description) < 10 {
		return errors.New("description must be at least 10 characters long")
	}

	if req.DiscountType != "percentage" && req.DiscountType != "fixed" {
		return errors.New("discount type is invalid")
	}
	switch req.DiscountType {

	case "percentage":
		if req.DiscountValue <= 0 || req.DiscountValue > 100 {
			return errors.New("percentage discount must be between 0 and 100")
		}
		if req.MaxDiscountAmount <= 0 {
			return errors.New("max discount amount required for percentage coupon")
		}

	case "fixed":
		if req.DiscountValue <= 0 {
			return errors.New("fixed discount must be greater than 0")
		}

		if req.MaxDiscountAmount > 0 {
			return errors.New("max discount amount should not be set for fixed coupons")
		}
	}

	if req.MinOrderAmount < 0 {
		return errors.New("min order amount cannot be negative")
	}

	if req.DiscountType == "fixed" && req.MinOrderAmount > 0 &&
		req.DiscountValue > req.MinOrderAmount {
		return errors.New("discount cannot exceed minimum order amount")
	}

	if req.ValidFrom.IsZero() || req.ValidTo.IsZero() {
		return errors.New("valid from and valid to are required")
	}
	if req.ValidFrom.After(req.ValidTo) {
		return errors.New("valid from cannot be after valid to")
	}

	if req.ValidTo.Before(time.Now()) {
		return errors.New("coupon already expired")
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
