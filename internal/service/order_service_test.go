package service

import (
	"context"
	"net/http"
	"testing"

	"github.com/razorpay/razorpay-go"
	"github.com/stretchr/testify/assert"
	"github.com/tharunn0/E-Commerce-Go/internal/apperror"
	"github.com/tharunn0/E-Commerce-Go/internal/config"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/order"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/shipping"
	
	cartmock "github.com/tharunn0/E-Commerce-Go/internal/domain/cart/mocks"
	ordermock "github.com/tharunn0/E-Commerce-Go/internal/domain/order/mocks"
	paymentmock "github.com/tharunn0/E-Commerce-Go/internal/domain/payment/mocks"
	productmock "github.com/tharunn0/E-Commerce-Go/internal/domain/product/mocks"
	promomock "github.com/tharunn0/E-Commerce-Go/internal/domain/promotion/mocks"
	usermock "github.com/tharunn0/E-Commerce-Go/internal/domain/user/mocks"
	
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestOrderService_UpdateOrderStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := usermock.NewMockUserRepository(ctrl)
	mockProductRepo := productmock.NewMockProductRepository(ctrl)
	mockCartRepo := cartmock.NewMockCartRepository(ctrl)
	mockOrderRepo := ordermock.NewMockOrderRepository(ctrl)
	mockPaymentRepo := paymentmock.NewMockPaymentRepository(ctrl)
	mockOfferRepo := promomock.NewMockOfferRepository(ctrl)
	mockCouponRepo := promomock.NewMockCouponRepository(ctrl)
	
	logger := zap.NewNop()
	cfg := config.OrderSettings{}
	razorpayClient := (*razorpay.Client)(nil) 

	orderService := NewOrderService(
		mockUserRepo, mockProductRepo, mockCartRepo, mockOrderRepo,
		mockPaymentRepo, mockOfferRepo, mockCouponRepo,
		razorpayClient, cfg, logger,
	)

	t.Run("fails when invalid status is provided", func(t *testing.T) {
		ctx := context.Background()
		orderID := "ORD-123"
		invalidStatus := shipping.ShipmentStatus("pending") 
		
		resp, err := orderService.UpdateOrderStatus(ctx, orderID, invalidStatus)
		
		assert.NotNil(t, err)
		assert.Equal(t, "INVALID_STATUS", err.Code)
		assert.Nil(t, resp)
	})
	
	t.Run("fails when order is not found in database", func(t *testing.T) {
		ctx := context.Background()
		orderID := "ORD-123"
		validStatus := shipping.ShipmentStatus("shipped") 
		
		mockOrderRepo.EXPECT().
			GetUserOrderByID(ctx, orderID, int64(0)).
			Return(nil, apperror.ErrOrderNotFound)
		
		resp, err := orderService.UpdateOrderStatus(ctx, orderID, validStatus)
		
		assert.NotNil(t, err)
		assert.Equal(t, http.StatusNotFound, err.Status)
		assert.Equal(t, apperror.ErrOrderNotFound.Error(), err.Message)
		assert.Nil(t, resp)
	})

	t.Run("fails when trying to update a cancelled order", func(t *testing.T) {
		ctx := context.Background()
		orderID := "ORD-456"
		validStatus := shipping.ShipmentStatus("shipped") 
		
		mockOrderRepo.EXPECT().
			GetUserOrderByID(ctx, orderID, int64(0)).
			Return(&order.OrderResponse{
				Status:         order.OrderStatusCancelled,
				ShipmentStatus: "", 
			}, nil)
			
		resp, err := orderService.UpdateOrderStatus(ctx, orderID, validStatus)
		
		assert.NotNil(t, err)
		assert.Equal(t, "INVALID_STATUS", err.Code)
		assert.Equal(t, "Order is already cancelled.", err.Message)
		assert.Nil(t, resp)
	})

	t.Run("successfully updates shipment status", func(t *testing.T) {
		ctx := context.Background()
		orderID := "ORD-456"
		validStatus := shipping.ShipmentStatus("shipped") 
		
		// 1. Service fetches the order to check current status
		mockOrderRepo.EXPECT().
			GetUserOrderByID(ctx, orderID, int64(0)).
			Return(&order.OrderResponse{
				Status:         "CONFIRMED",
				ShipmentStatus: "processing", 
				PaymentMethod:  "COD",
			}, nil)
			
		// 2. Service saves the new shipped status (and pass true because PaymentMethod is COD)
		mockOrderRepo.EXPECT().
			UpdateShipmentStatus(ctx, orderID, validStatus, true).
			Return(nil)
		
		resp, err := orderService.UpdateOrderStatus(ctx, orderID, validStatus)
		
		assert.Nil(t, err)
		assert.Nil(t, resp) // Note: this method returns nil, nil as seen in the codebase
	})
}
