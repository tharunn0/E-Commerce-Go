package service_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tharunn0/E-Commerce-Go/internal/config"
	"github.com/tharunn0/E-Commerce-Go/internal/domain"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/cart"
	cartMock "github.com/tharunn0/E-Commerce-Go/internal/domain/cart/mocks"
	productMock "github.com/tharunn0/E-Commerce-Go/internal/domain/product/mocks"
	offerMock "github.com/tharunn0/E-Commerce-Go/internal/domain/promotion/mocks"
	"github.com/tharunn0/E-Commerce-Go/internal/service"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func setupCartServiceTest(t *testing.T) (*cartMock.MockCartRepository, *productMock.MockProductRepository, *offerMock.MockOfferRepository, *service.CartService) {
	ctrl := gomock.NewController(t)
	mockCartRepo := cartMock.NewMockCartRepository(ctrl)
	mockProductRepo := productMock.NewMockProductRepository(ctrl)
	mockOfferRepo := offerMock.NewMockOfferRepository(ctrl)

	log := zap.NewNop()
	cartService := service.NewCartService(mockCartRepo, mockProductRepo, mockOfferRepo, config.CartSettings{MaxQuantity: 5}, log)

	return mockCartRepo, mockProductRepo, mockOfferRepo, cartService
}

func getMockContext(userID float64) context.Context {
	return context.WithValue(context.Background(), domain.KeyUserID, userID)
}

func TestCartService_AddToCart(t *testing.T) {
	mockCartRepo, _, _, cartService := setupCartServiceTest(t)
	ctx := getMockContext(1.0)

	req := &cart.AddToCartRequest{
		ProductVariantID: 10,
		Quantity:         2,
	}

	cartID := int64(100)
	expectedCart := &cart.Cart{
		Items: []*cart.CartItem{
			{
				ProductVariantID: 10,
				Quantity:         2,
				TotalPrice:       1000,
			},
		},
		CartTotalPrice: 1000,
	}

	mockCartRepo.EXPECT().
		AddToCart(gomock.Any(), int64(1), req.ProductVariantID, req.Quantity).
		Return(&cartID, nil).
		Times(1)

	mockCartRepo.EXPECT().
		GetCartByID(gomock.Any(), cartID).
		Return(expectedCart, nil).
		Times(1)

	result, apiErr := cartService.AddToCart(ctx, req)
	assert.Nil(t, apiErr)
	assert.NotNil(t, result)
	assert.Equal(t, expectedCart, result)
}

func TestCartService_AddToCart_InvalidQuantity(t *testing.T) {
	_, _, _, cartService := setupCartServiceTest(t)
	ctx := getMockContext(1.0)

	req := &cart.AddToCartRequest{
		ProductVariantID: 10,
		Quantity:         6,
	}

	result, apiErr := cartService.AddToCart(ctx, req)
	assert.Nil(t, result)
	assert.NotNil(t, apiErr)
	assert.Equal(t, "QUANTITY_EXCEEDED", apiErr.Code)
	assert.Equal(t, http.StatusUnprocessableEntity, apiErr.Status)
}

func TestCartService_UpdateCartItemQuantity(t *testing.T) {
	mockCartRepo, _, mockOfferRepo, cartService := setupCartServiceTest(t)
	ctx := getMockContext(1.0)

	req := &cart.UpdateCartItemQuantityRequest{
		ProductVariantID: 10,
		Quantity:         3,
	}

	updatedCartID := int64(100)
	expectedCart := &cart.Cart{
		Items: []*cart.CartItem{
			{
				ProductVariantID: 10,
				ProductID:        5,
				CategoryID:       2,
				Quantity:         3,
				OriginalPrice:    500,
				TotalPrice:       1500,
			},
		},
		CartTotalPrice: 1500,
	}

	mockCartRepo.EXPECT().
		UpdateCartItemQuantity(gomock.Any(), int64(1), req).
		Return(updatedCartID, nil).
		Times(1)

	mockCartRepo.EXPECT().
		GetCartByID(gomock.Any(), updatedCartID).
		Return(expectedCart, nil).
		Times(1)

	mockOfferRepo.EXPECT().
		GetAllActiveOffers(gomock.Any(), []int64{5}, []int64{2}).
		Return(nil, nil).
		Times(1)

	result, apiErr := cartService.UpdateCartItemQuantity(ctx, req)

	assert.Nil(t, apiErr)
	assert.NotNil(t, result)
	assert.Equal(t, expectedCart.CartTotalPrice, result.CartTotalPrice)
}
