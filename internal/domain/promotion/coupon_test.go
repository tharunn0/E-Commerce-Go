package promotion

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateCouponRequest_Validate(t *testing.T) {
	now := time.Now()
	tomorrow := now.Add(24 * time.Hour)

	tests := []struct {
		name    string
		req     *CreateCouponRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid Percentage Coupon",
			req: &CreateCouponRequest{
				CouponCode:        "SAVE10",
				Description:       "Get 10% off on your order",
				DiscountType:      "percentage",
				DiscountValue:     10,
				MaxDiscountAmount: 500,
				ValidFrom:         now,
				ValidTo:           tomorrow,
			},
			wantErr: false,
		},
		{
			name: "Valid Fixed Coupon",
			req: &CreateCouponRequest{
				CouponCode:     "SAVE100",
				Description:    "Get 100 off on your order",
				DiscountType:   "fixed",
				DiscountValue:  100,
				MinOrderAmount: 500,
				ValidFrom:      now,
				ValidTo:        tomorrow,
			},
			wantErr: false,
		},
		{
			name: "Invalid Coupon Code - Empty",
			req: &CreateCouponRequest{
				CouponCode: "",
			},
			wantErr: true,
			errMsg:  "coupon code is required",
		},
		{
			name: "Invalid Coupon Code - Short",
			req: &CreateCouponRequest{
				CouponCode: "S1",
			},
			wantErr: true,
			errMsg:  "coupon code must be between 3 and 10 characters long",
		},
		{
			name: "Invalid Coupon Code - Non Alphanumeric",
			req: &CreateCouponRequest{
				CouponCode: "SAVE-10",
			},
			wantErr: true,
			errMsg:  "coupon code must be uppercase alphanumeric",
		},
		{
			name: "Missing Description",
			req: &CreateCouponRequest{
				CouponCode: "SAVE10",
			},
			wantErr: true,
			errMsg:  "description is required",
		},
		{
			name: "Invalid Discount Type",
			req: &CreateCouponRequest{
				CouponCode:   "SAVE10",
				Description:  "Get 10% off on your order",
				DiscountType: "invalid",
			},
			wantErr: true,
			errMsg:  "discount type is invalid",
		},
		{
			name: "Percentage Discount Out of Range",
			req: &CreateCouponRequest{
				CouponCode:    "SAVE110",
				Description:   "Get 110% off on your order",
				DiscountType:  "percentage",
				DiscountValue: 110,
			},
			wantErr: true,
			errMsg:  "percentage discount must be between 0 and 100",
		},
		{
			name: "Missing Max Discount for Percentage",
			req: &CreateCouponRequest{
				CouponCode:    "SAVE10",
				Description:   "Get 10% off on your order",
				DiscountType:  "percentage",
				DiscountValue: 10,
			},
			wantErr: true,
			errMsg:  "max discount amount required for percentage coupon",
		},
		{
			name: "Max Discount Set for Fixed",
			req: &CreateCouponRequest{
				CouponCode:        "SAVE100",
				Description:       "Get 100 off on your order",
				DiscountType:      "fixed",
				DiscountValue:     100,
				MaxDiscountAmount: 500,
			},
			wantErr: true,
			errMsg:  "max discount amount should not be set for fixed coupons",
		},
		{
			name: "Invalid Dates",
			req: &CreateCouponRequest{
				CouponCode:        "SAVE10",
				Description:       "Get 10% off on your order",
				DiscountType:      "percentage",
				DiscountValue:     10,
				MaxDiscountAmount: 500,
				ValidFrom:         tomorrow,
				ValidTo:           now,
			},
			wantErr: true,
			errMsg:  "valid from cannot be after valid to",
		},
		{
			name: "Expired To Date",
			req: &CreateCouponRequest{
				CouponCode:        "SAVE10",
				Description:       "Get 10% off on your order",
				DiscountType:      "percentage",
				DiscountValue:     10,
				MaxDiscountAmount: 500,
				ValidFrom:         now.Add(-48 * time.Hour),
				ValidTo:           now.Add(-24 * time.Hour),
			},
			wantErr: true,
			errMsg:  "coupon already expired",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.errMsg, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateCoupon(t *testing.T) {
	now := time.Now()
	past := now.Add(-24 * time.Hour)
	future := now.Add(24 * time.Hour)

	tests := []struct {
		name    string
		coupon  *CouponResponse
		now     time.Time
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid Coupon",
			coupon: &CouponResponse{
				IsActive:  true,
				ValidFrom: past,
				ValidTo:   future,
			},
			now:     now,
			wantErr: false,
		},
		{
			name:    "Nil Coupon",
			coupon:  nil,
			now:     now,
			wantErr: true,
			errMsg:  "coupon not found",
		},
		{
			name: "Inactive Coupon",
			coupon: &CouponResponse{
				IsActive: false,
			},
			now:     now,
			wantErr: true,
			errMsg:  "coupon is not active",
		},
		{
			name: "Not Yet Valid",
			coupon: &CouponResponse{
				IsActive:  true,
				ValidFrom: future,
				ValidTo:   future.Add(24 * time.Hour),
			},
			now:     now,
			wantErr: true,
			errMsg:  "coupon is not yet valid",
		},
		{
			name: "Expired Coupon",
			coupon: &CouponResponse{
				IsActive:  true,
				ValidFrom: past.Add(-24 * time.Hour),
				ValidTo:   past,
			},
			now:     now,
			wantErr: true,
			errMsg:  "coupon is expired",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCoupon(tt.coupon, tt.now)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.errMsg, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
