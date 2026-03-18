package discount

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/cart"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/product"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/promotion"
)

func TestApplyDiscounts(t *testing.T) {
	offers := []*promotion.Offer{
		{
			ID:            1,
			Name:          "Product 10% Off",
			DiscountType:  "percentage",
			DiscountValue: 10,
			ProductIDs:    []int64{1},
		},
		{
			ID:            2,
			Name:          "Category 100 Off",
			DiscountType:  "flat",
			DiscountValue: 100,
			CategoryIDs:   []int64{10},
		},
		{
			ID:            3,
			Name:          "Flat Discount exceeding total",
			DiscountType:  "flat",
			DiscountValue: 2000,
			ProductIDs:    []int64{3},
		},
		{
			ID:            4,
			Name:          "Invalid Discount Type",
			DiscountType:  "invalid",
			DiscountValue: 10,
			ProductIDs:    []int64{4},
		},
	}

	tests := []struct {
		name               string
		cart               *cart.Cart
		expectedTotalPrice float64
		expectedItemPrices []float64
	}{
		{
			name: "Discount on product",
			cart: &cart.Cart{
				Items: []*cart.CartItem{
					{
						ProductID:     1,
						OriginalPrice: 1000,
						Quantity:      1,
					},
				},
			},
			expectedTotalPrice: 900,
			expectedItemPrices: []float64{900},
		},
		{
			name: "Discount on category",
			cart: &cart.Cart{
				Items: []*cart.CartItem{
					{
						CategoryID:    10,
						OriginalPrice: 1000,
						Quantity:      1,
					},
				},
			},
			expectedTotalPrice: 900,
			expectedItemPrices: []float64{900},
		},
		{
			name: "Multiple items, one discounted",
			cart: &cart.Cart{
				Items: []*cart.CartItem{
					{
						ProductID:     1,
						OriginalPrice: 1000,
						Quantity:      1,
					},
					{
						ProductID:     2,
						OriginalPrice: 500,
						Quantity:      1,
					},
				},
			},
			expectedTotalPrice: 1400,
			expectedItemPrices: []float64{900, 500},
		},
		{
			name: "Quantity check",
			cart: &cart.Cart{
				Items: []*cart.CartItem{
					{
						ProductID:     1,
						OriginalPrice: 1000,
						Quantity:      2,
					},
				},
			},
			expectedTotalPrice: 1800,
			expectedItemPrices: []float64{1800},
		},
		{
			name: "Discount exceeding base total",
			cart: &cart.Cart{
				Items: []*cart.CartItem{
					{
						ProductID:     3,
						OriginalPrice: 1000,
						Quantity:      1,
					},
				},
			},
			expectedTotalPrice: 0,
			expectedItemPrices: []float64{0},
		},
		{
			name: "Invalid discount type",
			cart: &cart.Cart{
				Items: []*cart.CartItem{
					{
						ProductID:     4,
						OriginalPrice: 1000,
						Quantity:      1,
					},
				},
			},
			expectedTotalPrice: 1000,
			expectedItemPrices: []float64{1000},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ApplyDiscounts(tt.cart, offers)
			assert.Equal(t, tt.expectedTotalPrice, tt.cart.CartTotalPrice)
			for i, item := range tt.cart.Items {
				assert.Equal(t, tt.expectedItemPrices[i], item.TotalPrice)
			}
		})
	}
}

func TestApplyDiscountsToVariants(t *testing.T) {
	offers := []*promotion.Offer{
		{
			ID:            1,
			Name:          "Product 10% Off",
			DiscountType:  "percentage",
			DiscountValue: 10,
			ProductIDs:    []int64{1},
		},
		{
			ID:            2,
			Name:          "Category Flat Off",
			DiscountType:  "flat",
			DiscountValue: 200,
			CategoryIDs:   []int64{10},
		},
		{
			ID:            3,
			Name:          "Product Flat Over Base Total Off",
			DiscountType:  "flat",
			DiscountValue: 1500, // original price is 1000, so capping should occur
			ProductIDs:    []int64{3},
		},
		{
			ID:            4,
			Name:          "Invalid Discount Type",
			DiscountType:  "invalid",
			DiscountValue: 100,
			ProductIDs:    []int64{4},
		},
	}

	tests := []struct {
		name              string
		variants          []*product.ProductVariantResponse
		expectedSalePrice *float64
	}{
		{
			name: "Variant with matching product offer - percentage",
			variants: []*product.ProductVariantResponse{
				{
					BaseProduct:   &product.BaseProduct{ID: 1},
					OriginalPrice: 1000,
				},
			},
			expectedSalePrice: func() *float64 { f := 900.0; return &f }(),
		},
		{
			name: "Variant with matching category offer - flat",
			variants: []*product.ProductVariantResponse{
				{
					BaseProduct:   &product.BaseProduct{ID: 2, CategoryID: 10},
					OriginalPrice: 1000,
				},
			},
			expectedSalePrice: func() *float64 { f := 800.0; return &f }(),
		},
		{
			name: "Variant discount capped",
			variants: []*product.ProductVariantResponse{
				{
					BaseProduct:   &product.BaseProduct{ID: 3},
					OriginalPrice: 1000,
				},
			},
			expectedSalePrice: func() *float64 { f := 0.0; return &f }(),
		},
		{
			name: "Variant invalid discount type",
			variants: []*product.ProductVariantResponse{
				{
					BaseProduct:   &product.BaseProduct{ID: 4},
					OriginalPrice: 1000,
				},
			},
			expectedSalePrice: nil,
		},
		{
			name: "Variant with non-matching product",
			variants: []*product.ProductVariantResponse{
				{
					BaseProduct:   &product.BaseProduct{ID: 2},
					OriginalPrice: 1000,
				},
			},
			expectedSalePrice: nil,
		},
		{
			name: "Nil variant and base product",
			variants: []*product.ProductVariantResponse{
				nil,
				{
					BaseProduct: nil,
				},
			},
			expectedSalePrice: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ApplyDiscountsToVariants(tt.variants, offers)
			if tt.name == "Nil variant and base product" {
				return // we skip assertion for this
			}
			if tt.expectedSalePrice == nil {
				assert.Nil(t, tt.variants[0].SalePrice)
			} else {
				assert.NotNil(t, tt.variants[0].SalePrice)
				assert.Equal(t, *tt.expectedSalePrice, *tt.variants[0].SalePrice)
			}
		})
	}
}

func TestContainsId(t *testing.T) {
	assert.True(t, containsId([]int64{1, 2, 3}, 2))
	assert.False(t, containsId([]int64{1, 2, 3}, 4))
	assert.False(t, containsId([]int64{}, 1))
}
