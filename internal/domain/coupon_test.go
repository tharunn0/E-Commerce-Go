package domain

import (
	"testing"
	"time"
)

func TestCreateCouponRequest_Validate(t *testing.T) {
	validBase := func() *CreateCouponRequest {
		return &CreateCouponRequest{
			CouponCode:        "SAVE10",
			Description:       "Get 10% off your order",
			DiscountType:      "percentage",
			DiscountValue:     10,
			MinOrderAmount:    100,
			MaxDiscountAmount: 50,
			ValidFrom:         time.Now(),
			ValidTo:           time.Now().Add(24 * time.Hour),
		}
	}

	tests := []struct {
		name    string
		mutate  func(*CreateCouponRequest)
		wantErr string
	}{
		{
			name:    "valid request",
			mutate:  func(r *CreateCouponRequest) {},
			wantErr: "",
		},
		{
			name:    "missing coupon code",
			mutate:  func(r *CreateCouponRequest) { r.CouponCode = "" },
			wantErr: "coupon code is required",
		},
		{
			name:    "coupon code too short",
			mutate:  func(r *CreateCouponRequest) { r.CouponCode = "AB" },
			wantErr: "coupon code must be between 3 and 10 characters long",
		},
		{
			name:    "coupon code too long",
			mutate:  func(r *CreateCouponRequest) { r.CouponCode = "TOOLONGCODE1" },
			wantErr: "coupon code must be between 3 and 10 characters long",
		},
		{
			name:    "missing description",
			mutate:  func(r *CreateCouponRequest) { r.Description = "" },
			wantErr: "description is required",
		},
		{
			name:    "description too short",
			mutate:  func(r *CreateCouponRequest) { r.Description = "Short" },
			wantErr: "description must be at least 10 characters long",
		},
		{
			name:    "invalid discount type",
			mutate:  func(r *CreateCouponRequest) { r.DiscountType = "bogo" },
			wantErr: "discount type is invalid",
		},
		{
			name:    "fixed discount type is valid",
			mutate:  func(r *CreateCouponRequest) { r.DiscountType = "fixed" },
			wantErr: "",
		},
		{
			name:    "zero discount value",
			mutate:  func(r *CreateCouponRequest) { r.DiscountValue = 0 },
			wantErr: "invalid discount value",
		},
		{
			name:    "negative discount value",
			mutate:  func(r *CreateCouponRequest) { r.DiscountValue = -5 },
			wantErr: "invalid discount value",
		},
		{
			name:    "zero min order amount",
			mutate:  func(r *CreateCouponRequest) { r.MinOrderAmount = 0 },
			wantErr: "invalid min order amount",
		},
		{
			name:    "zero max discount amount",
			mutate:  func(r *CreateCouponRequest) { r.MaxDiscountAmount = 0 },
			wantErr: "invalid max discount amount",
		},
		{
			name:    "zero valid from",
			mutate:  func(r *CreateCouponRequest) { r.ValidFrom = time.Time{} },
			wantErr: "invalid valid from",
		},
		{
			name:    "zero valid to",
			mutate:  func(r *CreateCouponRequest) { r.ValidTo = time.Time{} },
			wantErr: "invalid valid to",
		},
		{
			name: "valid from after valid to",
			mutate: func(r *CreateCouponRequest) {
				r.ValidFrom = time.Now().Add(48 * time.Hour)
				r.ValidTo = time.Now().Add(24 * time.Hour)
			},
			wantErr: "valid from cannot be after valid to",
		},
		{
			name: "valid from equal to valid to",
			mutate: func(r *CreateCouponRequest) {
				t := time.Now().Add(24 * time.Hour)
				r.ValidFrom = t
				r.ValidTo = t
			},
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := validBase()
			tt.mutate(req)
			err := req.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("expected error %q, got nil", tt.wantErr)
				} else if err.Error() != tt.wantErr {
					t.Errorf("expected error %q, got %q", tt.wantErr, err.Error())
				}
			}
		})
	}
}

func TestValidateCoupon(t *testing.T) {
	now := time.Now()

	validCoupon := func() *CouponResponse {
		return &CouponResponse{
			CouponCode: "SAVE10",
			IsActive:   true,
			ValidFrom:  now.Add(-time.Hour),
			ValidTo:    now.Add(time.Hour),
		}
	}

	tests := []struct {
		name    string
		coupon  *CouponResponse
		now     time.Time
		wantErr string
	}{
		{
			name:    "valid coupon",
			coupon:  validCoupon(),
			now:     now,
			wantErr: "",
		},
		{
			name:    "nil coupon",
			coupon:  nil,
			now:     now,
			wantErr: "coupon not found",
		},
		{
			name: "inactive coupon",
			coupon: func() *CouponResponse {
				c := validCoupon()
				c.IsActive = false
				return c
			}(),
			now:     now,
			wantErr: "coupon is not active",
		},
		{
			name: "coupon not yet valid",
			coupon: func() *CouponResponse {
				c := validCoupon()
				c.ValidFrom = now.Add(time.Hour)
				return c
			}(),
			now:     now,
			wantErr: "coupon is not yet valid",
		},
		{
			name: "coupon expired",
			coupon: func() *CouponResponse {
				c := validCoupon()
				c.ValidTo = now.Add(-time.Hour)
				return c
			}(),
			now:     now,
			wantErr: "coupon is expired",
		},
		{
			name: "coupon valid exactly at ValidFrom boundary",
			coupon: func() *CouponResponse {
				c := validCoupon()
				c.ValidFrom = now
				return c
			}(),
			now:     now,
			wantErr: "",
		},
		{
			name: "coupon valid exactly at ValidTo boundary",
			coupon: func() *CouponResponse {
				c := validCoupon()
				c.ValidTo = now
				return c
			}(),
			now:     now,
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCoupon(tt.coupon, tt.now)
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("expected error %q, got nil", tt.wantErr)
				} else if err.Error() != tt.wantErr {
					t.Errorf("expected error %q, got %q", tt.wantErr, err.Error())
				}
			}
		})
	}
}

func TestApplyCouponToCart(t *testing.T) {
	salePrice := 80.0

	validCart := func() *Cart {
		return &Cart{
			CartTotalPrice: 500.0,
		}
	}

	validCoupon := func() *CouponResponse {
		return &CouponResponse{
			CouponCode:        "SAVE10",
			DiscountType:      "percentage",
			DiscountValue:     10,
			MaxDiscountAmount: 100,
		}
	}

	_ = salePrice

	tests := []struct {
		name         string
		cart         *Cart
		coupon       *CouponResponse
		wantErr      string
		wantTotal    float64
		wantDiscount float64
		wantMessage  string
	}{
		{
			name:    "nil cart",
			cart:    nil,
			coupon:  validCoupon(),
			wantErr: "invalid input",
		},
		{
			name:    "nil coupon",
			cart:    validCart(),
			coupon:  nil,
			wantErr: "invalid input",
		},
		{
			name:    "invalid discount type",
			cart:    validCart(),
			coupon:  func() *CouponResponse { c := validCoupon(); c.DiscountType = "bogo"; return c }(),
			wantErr: "invalid discount type",
		},
		{
			name: "percentage discount within max cap",
			cart: validCart(), // total: 500
			coupon: func() *CouponResponse {
				c := validCoupon()
				c.DiscountType = "percentage"
				c.DiscountValue = 10      // 10% of 500 = 50
				c.MaxDiscountAmount = 100 // 50 < 100, no cap
				return c
			}(),
			wantErr:      "",
			wantTotal:    450.0,
			wantDiscount: 50.0,
			wantMessage:  "Discount applied successfully",
		},
		{
			name: "percentage discount exceeds max cap",
			cart: validCart(), // total: 500
			coupon: func() *CouponResponse {
				c := validCoupon()
				c.DiscountType = "percentage"
				c.DiscountValue = 30      // 30% of 500 = 150
				c.MaxDiscountAmount = 100 // 150 > 100, capped at 100
				return c
			}(),
			wantErr:      "",
			wantTotal:    400.0,
			wantDiscount: 100.0,
			wantMessage:  "Maximum discount applied",
		},
		{
			name: "fixed discount within cart total",
			cart: validCart(), // total: 500
			coupon: func() *CouponResponse {
				c := validCoupon()
				c.DiscountType = "fixed"
				c.DiscountValue = 200
				c.MaxDiscountAmount = 300
				return c
			}(),
			wantErr:      "",
			wantTotal:    300.0,
			wantDiscount: 200.0,
			wantMessage:  "Discount applied successfully",
		},
		{
			name: "fixed discount exceeds max cap",
			cart: validCart(), // total: 500
			coupon: func() *CouponResponse {
				c := validCoupon()
				c.DiscountType = "fixed"
				c.DiscountValue = 400
				c.MaxDiscountAmount = 150
				return c
			}(),
			wantErr:      "",
			wantTotal:    350.0,
			wantDiscount: 150.0,
			wantMessage:  "Maximum discount applied",
		},
		{
			name: "discount exceeds cart total",
			cart: &Cart{CartTotalPrice: 50.0},
			coupon: func() *CouponResponse {
				c := validCoupon()
				c.DiscountType = "fixed"
				c.DiscountValue = 200
				c.MaxDiscountAmount = 500
				return c
			}(),
			wantErr:      "",
			wantTotal:    0.0,
			wantDiscount: 50.0,
			wantMessage:  "Discount is greater than cart total",
		},
		{
			name: "coupon data populated correctly",
			cart: validCart(),
			coupon: func() *CouponResponse {
				c := validCoupon()
				c.CouponCode = "FLAT50"
				c.DiscountType = "fixed"
				c.DiscountValue = 50
				c.MaxDiscountAmount = 200
				return c
			}(),
			wantErr:      "",
			wantTotal:    450.0,
			wantDiscount: 50.0,
			wantMessage:  "Discount applied successfully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ApplyCouponToCart(tt.cart, tt.coupon)

			if tt.wantErr != "" {
				if err == nil {
					t.Errorf("expected error %q, got nil", tt.wantErr)
				} else if err.Error() != tt.wantErr {
					t.Errorf("expected error %q, got %q", tt.wantErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.cart.CartTotalPrice != tt.wantTotal {
				t.Errorf("cart total: want %.2f, got %.2f", tt.wantTotal, tt.cart.CartTotalPrice)
			}

			cd := tt.cart.CouponData
			if cd == nil {
				t.Fatal("expected CouponData to be set, got nil")
			}
			if cd.DiscountedAmount != tt.wantDiscount {
				t.Errorf("discounted amount: want %.2f, got %.2f", tt.wantDiscount, cd.DiscountedAmount)
			}
			if cd.Message != tt.wantMessage {
				t.Errorf("message: want %q, got %q", tt.wantMessage, cd.Message)
			}
			if tt.coupon != nil && cd.CouponCode != tt.coupon.CouponCode {
				t.Errorf("coupon code: want %q, got %q", tt.coupon.CouponCode, cd.CouponCode)
			}
		})
	}
}
