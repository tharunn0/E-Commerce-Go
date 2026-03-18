package cart

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/promotion"
)

func TestApplyCouponToCart(t *testing.T) {
	tests := []struct {
		name          string
		cart          *Cart
		coupon        *promotion.CouponResponse
		expectedError error
		expectedTotal float64
		expectedMsg   string
	}{
		{
			name: "Successful Percentage Discount",
			cart: &Cart{
				CartTotalPrice: 1000,
			},
			coupon: &promotion.CouponResponse{
				CouponCode:    "SAVE10",
				DiscountType:  "percentage",
				DiscountValue: 10,
			},
			expectedError: nil,
			expectedTotal: 900,
			expectedMsg:   "Discount applied successfully",
		},
		{
			name: "Successful Fixed Discount",
			cart: &Cart{
				CartTotalPrice: 1000,
			},
			coupon: &promotion.CouponResponse{
				CouponCode:    "SAVE100",
				DiscountType:  "fixed",
				DiscountValue: 100,
			},
			expectedError: nil,
			expectedTotal: 900,
			expectedMsg:   "Discount applied successfully",
		},
		{
			name: "Percentage Discount with Max Cap",
			cart: &Cart{
				CartTotalPrice: 2000,
			},
			coupon: &promotion.CouponResponse{
				CouponCode:        "SAVE50",
				DiscountType:      "percentage",
				DiscountValue:     50,
				MaxDiscountAmount: 500,
			},
			expectedError: nil,
			expectedTotal: 1500,
			expectedMsg:   "Maximum discount applied",
		},
		{
			name: "Discount Greater than Cart Total",
			cart: &Cart{
				CartTotalPrice: 50,
			},
			coupon: &promotion.CouponResponse{
				CouponCode:    "SAVE100",
				DiscountType:  "fixed",
				DiscountValue: 100,
			},
			expectedError: nil,
			expectedTotal: 0,
			expectedMsg:   "Discount is greater than cart total",
		},
		{
			name: "Invalid Discount Type",
			cart: &Cart{
				CartTotalPrice: 1000,
			},
			coupon: &promotion.CouponResponse{
				CouponCode:   "INVALID",
				DiscountType: "invalid",
			},
			expectedError: errors.New("invalid discount type"),
		},
		{
			name: "Nil Input",
			cart: nil,
			coupon: &promotion.CouponResponse{
				CouponCode: "NIL",
			},
			expectedError: errors.New("invalid input"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ApplyCouponToCart(tt.cart, tt.coupon)
			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedTotal, tt.cart.CartTotalPrice)
				assert.Equal(t, tt.expectedMsg, tt.cart.CouponData.Message)
				assert.Equal(t, tt.coupon.CouponCode, tt.cart.CouponData.CouponCode)
			}
		})
	}
}
