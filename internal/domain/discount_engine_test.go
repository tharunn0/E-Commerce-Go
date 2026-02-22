package domain

import (
	"testing"
)

func TestApplyDiscounts(t *testing.T) {
	validCart := func() *Cart {
		return &Cart{
			Items: []*CartItem{
				{
					ProductID:     1,
					CategoryID:    10,
					Quantity:      2,
					OriginalPrice: 100.0, // baseTotal = 200
				},
			},
		}
	}

	tests := []struct {
		name             string
		cart             *Cart
		offers           []*Offer
		wantCartTotal    float64
		wantItemTotal    float64
		wantSalePrice    *float64
		wantDiscount     float64
		wantAppliedOffer bool
	}{
		{
			name:          "no offers",
			cart:          validCart(),
			offers:        []*Offer{},
			wantCartTotal: 200.0,
			wantItemTotal: 200.0,
			wantSalePrice: nil,
		},
		{
			name: "percentage offer by product id",
			cart: validCart(),
			offers: []*Offer{
				{
					ID:            1,
					Name:          "10% off",
					DiscountType:  "percentage",
					DiscountValue: 10, // 10% of 200 = 20
					ProductIDs:    []int64{1},
				},
			},
			wantCartTotal:    180.0,
			wantItemTotal:    180.0,
			wantSalePrice:    float64Ptr(90.0),
			wantDiscount:     20.0,
			wantAppliedOffer: true,
		},
		{
			name: "percentage offer by category id",
			cart: validCart(),
			offers: []*Offer{
				{
					ID:            2,
					Name:          "20% off category",
					DiscountType:  "percentage",
					DiscountValue: 20, // 20% of 200 = 40
					CategoryIDs:   []int64{10},
				},
			},
			wantCartTotal:    160.0,
			wantItemTotal:    160.0,
			wantSalePrice:    float64Ptr(80.0),
			wantDiscount:     40.0,
			wantAppliedOffer: true,
		},
		{
			name: "flat offer by product id",
			cart: validCart(),
			offers: []*Offer{
				{
					ID:            3,
					Name:          "flat 50 off",
					DiscountType:  "flat",
					DiscountValue: 50,
					ProductIDs:    []int64{1},
				},
			},
			wantCartTotal:    150.0,
			wantItemTotal:    150.0,
			wantSalePrice:    float64Ptr(75.0),
			wantDiscount:     50.0,
			wantAppliedOffer: true,
		},
		{
			name: "best discount selected among multiple offers",
			cart: validCart(),
			offers: []*Offer{
				{
					ID:            1,
					Name:          "10% off",
					DiscountType:  "percentage",
					DiscountValue: 10, // 20
					ProductIDs:    []int64{1},
				},
				{
					ID:            2,
					Name:          "flat 60 off",
					DiscountType:  "flat",
					DiscountValue: 60, // 60 — better
					ProductIDs:    []int64{1},
				},
			},
			wantCartTotal:    140.0,
			wantItemTotal:    140.0,
			wantSalePrice:    float64Ptr(70.0),
			wantDiscount:     60.0,
			wantAppliedOffer: true,
		},
		{
			name: "offer does not match product or category",
			cart: validCart(),
			offers: []*Offer{
				{
					ID:            1,
					DiscountType:  "percentage",
					DiscountValue: 10,
					ProductIDs:    []int64{99},
					CategoryIDs:   []int64{99},
				},
			},
			wantCartTotal: 200.0,
			wantItemTotal: 200.0,
			wantSalePrice: nil,
		},
		{
			name: "invalid discount type is skipped",
			cart: validCart(),
			offers: []*Offer{
				{
					ID:            1,
					DiscountType:  "bogo",
					DiscountValue: 50,
					ProductIDs:    []int64{1},
				},
			},
			wantCartTotal: 200.0,
			wantItemTotal: 200.0,
			wantSalePrice: nil,
		},
		{
			name: "discount capped at base total",
			cart: validCart(), // baseTotal = 200
			offers: []*Offer{
				{
					ID:            1,
					DiscountType:  "flat",
					DiscountValue: 500, // exceeds 200, capped at 200
					ProductIDs:    []int64{1},
				},
			},
			wantCartTotal:    0.0,
			wantItemTotal:    0.0,
			wantSalePrice:    float64Ptr(0.0),
			wantDiscount:     200.0,
			wantAppliedOffer: true,
		},
		{
			name: "multiple items cart total accumulated correctly",
			cart: &Cart{
				Items: []*CartItem{
					{ProductID: 1, CategoryID: 10, Quantity: 2, OriginalPrice: 100.0}, // 200
					{ProductID: 2, CategoryID: 20, Quantity: 1, OriginalPrice: 50.0},  // 50, no offer
				},
			},
			offers: []*Offer{
				{
					ID:            1,
					DiscountType:  "percentage",
					DiscountValue: 10, // 10% of 200 = 20
					ProductIDs:    []int64{1},
				},
			},
			wantCartTotal:    230.0, // 180 + 50
			wantItemTotal:    180.0,
			wantSalePrice:    float64Ptr(90.0),
			wantAppliedOffer: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ApplyDiscounts(tt.cart, tt.offers)

			if tt.cart.CartTotalPrice != tt.wantCartTotal {
				t.Errorf("cart total: want %.2f, got %.2f", tt.wantCartTotal, tt.cart.CartTotalPrice)
			}

			firstItem := tt.cart.Items[0]
			if firstItem.TotalPrice != tt.wantItemTotal {
				t.Errorf("item total: want %.2f, got %.2f", tt.wantItemTotal, firstItem.TotalPrice)
			}

			if tt.wantSalePrice == nil {
				if firstItem.SalePrice != nil {
					t.Errorf("sale price: want nil, got %.2f", *firstItem.SalePrice)
				}
			} else {
				if firstItem.SalePrice == nil {
					t.Errorf("sale price: want %.2f, got nil", *tt.wantSalePrice)
				} else if *firstItem.SalePrice != *tt.wantSalePrice {
					t.Errorf("sale price: want %.2f, got %.2f", *tt.wantSalePrice, *firstItem.SalePrice)
				}
			}

			if tt.wantAppliedOffer && firstItem.AppliedOffer == nil {
				t.Errorf("expected AppliedOffer to be set, got nil")
			}
			if !tt.wantAppliedOffer && firstItem.AppliedOffer != nil {
				t.Errorf("expected AppliedOffer to be nil, got %+v", firstItem.AppliedOffer)
			}
		})
	}
}

func TestApplyDiscountsToVariants(t *testing.T) {
	validVariant := func() *ProductVariantResponse {
		return &ProductVariantResponse{
			OriginalPrice: 200.0,
			BaseProduct: &BaseProduct{
				ID:         1,
				CategoryID: 10,
			},
		}
	}

	tests := []struct {
		name             string
		variants         []*ProductVariantResponse
		offers           []*Offer
		wantSalePrice    *float64
		wantDiscount     float64
		wantAppliedOffer bool
	}{
		{
			name:          "no offers",
			variants:      []*ProductVariantResponse{validVariant()},
			offers:        []*Offer{},
			wantSalePrice: nil,
		},
		{
			name:     "nil variant skipped",
			variants: []*ProductVariantResponse{nil},
			offers: []*Offer{
				{DiscountType: "percentage", DiscountValue: 10, ProductIDs: []int64{1}},
			},
			wantSalePrice: nil,
		},
		{
			name: "variant with nil base product skipped",
			variants: []*ProductVariantResponse{
				{OriginalPrice: 200.0, BaseProduct: nil},
			},
			offers: []*Offer{
				{DiscountType: "percentage", DiscountValue: 10, ProductIDs: []int64{1}},
			},
			wantSalePrice: nil,
		},
		{
			name:     "percentage offer by product id",
			variants: []*ProductVariantResponse{validVariant()},
			offers: []*Offer{
				{
					ID:            1,
					Name:          "10% off",
					DiscountType:  "percentage",
					DiscountValue: 10, // 10% of 200 = 20
					ProductIDs:    []int64{1},
				},
			},
			wantSalePrice:    float64Ptr(180.0),
			wantDiscount:     20.0,
			wantAppliedOffer: true,
		},
		{
			name:     "percentage offer by category id",
			variants: []*ProductVariantResponse{validVariant()},
			offers: []*Offer{
				{
					ID:            2,
					Name:          "20% off category",
					DiscountType:  "percentage",
					DiscountValue: 20, // 20% of 200 = 40
					CategoryIDs:   []int64{10},
				},
			},
			wantSalePrice:    float64Ptr(160.0),
			wantDiscount:     40.0,
			wantAppliedOffer: true,
		},
		{
			name:     "flat offer applied",
			variants: []*ProductVariantResponse{validVariant()},
			offers: []*Offer{
				{
					ID:            3,
					Name:          "flat 50 off",
					DiscountType:  "flat",
					DiscountValue: 50,
					ProductIDs:    []int64{1},
				},
			},
			wantSalePrice:    float64Ptr(150.0),
			wantDiscount:     50.0,
			wantAppliedOffer: true,
		},
		{
			name:     "best offer selected among multiple",
			variants: []*ProductVariantResponse{validVariant()},
			offers: []*Offer{
				{
					ID:            1,
					DiscountType:  "percentage",
					DiscountValue: 10, // 20
					ProductIDs:    []int64{1},
				},
				{
					ID:            2,
					DiscountType:  "flat",
					DiscountValue: 80, // 80 — better
					ProductIDs:    []int64{1},
				},
			},
			wantSalePrice:    float64Ptr(120.0),
			wantDiscount:     80.0,
			wantAppliedOffer: true,
		},
		{
			name:     "offer does not match",
			variants: []*ProductVariantResponse{validVariant()},
			offers: []*Offer{
				{
					DiscountType:  "percentage",
					DiscountValue: 10,
					ProductIDs:    []int64{99},
					CategoryIDs:   []int64{99},
				},
			},
			wantSalePrice: nil,
		},
		{
			name:     "invalid discount type skipped",
			variants: []*ProductVariantResponse{validVariant()},
			offers: []*Offer{
				{DiscountType: "bogo", DiscountValue: 50, ProductIDs: []int64{1}},
			},
			wantSalePrice: nil,
		},
		{
			name:     "discount capped at original price",
			variants: []*ProductVariantResponse{validVariant()}, // 200
			offers: []*Offer{
				{
					ID:            1,
					DiscountType:  "flat",
					DiscountValue: 500, // exceeds 200, capped
					ProductIDs:    []int64{1},
				},
			},
			wantSalePrice:    float64Ptr(0.0),
			wantDiscount:     200.0,
			wantAppliedOffer: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ApplyDiscountsToVariants(tt.variants, tt.offers)

			v := tt.variants[0]
			if v == nil || v.BaseProduct == nil {
				return // skipped variants, nothing to assert
			}

			if tt.wantSalePrice == nil {
				if v.SalePrice != nil {
					t.Errorf("sale price: want nil, got %.2f", *v.SalePrice)
				}
			} else {
				if v.SalePrice == nil {
					t.Errorf("sale price: want %.2f, got nil", *tt.wantSalePrice)
				} else if *v.SalePrice != *tt.wantSalePrice {
					t.Errorf("sale price: want %.2f, got %.2f", *tt.wantSalePrice, *v.SalePrice)
				}
			}

			if tt.wantAppliedOffer && v.AppliedOffer == nil {
				t.Errorf("expected AppliedOffer to be set, got nil")
			}
			if !tt.wantAppliedOffer && v.AppliedOffer != nil {
				t.Errorf("expected AppliedOffer to be nil, got %+v", v.AppliedOffer)
			}
			if tt.wantAppliedOffer && v.AppliedOffer != nil {
				if v.AppliedOffer.DiscountAmount != tt.wantDiscount {
					t.Errorf("discount amount: want %.2f, got %.2f", tt.wantDiscount, v.AppliedOffer.DiscountAmount)
				}
			}
		})
	}
}

// helper
func float64Ptr(f float64) *float64 {
	return &f
}
