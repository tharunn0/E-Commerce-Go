package payment_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/payment"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/payment/mocks"
	"go.uber.org/mock/gomock"
)

type ExampleService struct {
	gateway payment.PaymentGateway
}

func (s *ExampleService) ProcessOrder(ctx context.Context, userID int64, amount int64) (*payment.Response, error) {
	req := payment.Request{
		OrderID:  "ORDER123",
		UserID:   userID,
		Amount:   amount,
		Currency: "INR",
	}
	return s.gateway.CreatePayment(ctx, req)
}

func TestProcessOrder_WithMockGen(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGateway := mocks.NewMockPaymentGateway(ctrl)

	expectedReq := payment.Request{
		OrderID:  "ORDER123",
		UserID:   1,
		Amount:   500,
		Currency: "INR",
	}

	mockResponse := &payment.Response{
		Status:   payment.PaymentStatusPending,
		Provider: payment.PaymentMethodRazorpay,
	}

	mockGateway.EXPECT().
		CreatePayment(gomock.Any(), expectedReq).
		Return(mockResponse, nil).
		Times(1)
	service := ExampleService{
		gateway: mockGateway,
	}
	resp, err := service.ProcessOrder(context.Background(), 1, 500)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, payment.PaymentStatusPending, resp.Status)
}
