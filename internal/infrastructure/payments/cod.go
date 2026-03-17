package payments

import (
	"context"

	"github.com/tharunn0/E-Commerce-Go/internal/domain/payment"
)

type CODPayment struct{}

func (c *CODPayment) CreatePayment(ctx context.Context, req payment.Request) (*payment.Response, error) {
	return &payment.Response{
		Status:     payment.PaymentStatusPending,
		Provider:   payment.PaymentMethodCOD,
		GatewayRef: nil,
		PaymentURL: nil,
	}, nil
}
