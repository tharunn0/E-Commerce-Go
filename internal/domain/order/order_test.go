package order

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/payment"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/shipping"
)

func TestCreateOrderRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     *CreateOrderRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid Request COD",
			req: &CreateOrderRequest{
				AddressID:     1,
				DeliveryType:  shipping.DeliveryTypeNormal,
				PaymentMethod: payment.PaymentMethodCOD,
			},
			wantErr: false,
		},
		{
			name: "Valid Request Razorpay",
			req: &CreateOrderRequest{
				AddressID:     1,
				DeliveryType:  shipping.DeliveryTypeExpress,
				PaymentMethod: payment.PaymentMethodRazorpay,
			},
			wantErr: false,
		},
		{
			name: "Invalid Delivery Type",
			req: &CreateOrderRequest{
				AddressID:     1,
				DeliveryType:  "INVALID",
				PaymentMethod: payment.PaymentMethodCOD,
			},
			wantErr: true,
			errMsg:  "invalid delivery type",
		},
		{
			name: "Invalid Address ID",
			req: &CreateOrderRequest{
				AddressID:     0,
				DeliveryType:  shipping.DeliveryTypeNormal,
				PaymentMethod: payment.PaymentMethodCOD,
			},
			wantErr: true,
			errMsg:  "invalid address",
		},
		{
			name: "Invalid Payment Method",
			req: &CreateOrderRequest{
				AddressID:     1,
				DeliveryType:  shipping.DeliveryTypeNormal,
				PaymentMethod: "INVALID",
			},
			wantErr: true,
			errMsg:  "invalid payment method",
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

func TestOrderStatusFlow(t *testing.T) {
	assert.NotNil(t, OrderStatusFlow)
	assert.Contains(t, OrderStatusFlow[OrderStatusPending], OrderStatusConfirmed)
	assert.Contains(t, OrderStatusFlow[OrderStatusConfirmed], OrderStatusShipped)
	assert.Contains(t, OrderStatusFlow[OrderStatusShipped], OrderStatusDelivered)
	assert.Empty(t, OrderStatusFlow[OrderStatusCancelled])
	assert.Empty(t, OrderStatusFlow[OrderStatusDelivered])
}
