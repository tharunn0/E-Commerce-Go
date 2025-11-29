package payments

import (
	"context"

	"github.com/tharunn0/E-Commerce-Go/internal/domain"
)

type CODPayment struct{}

func (c *CODPayment) CreatePayment(ctx context.Context, req domain.PaymentRequest) (*domain.PaymentResponse, error) {
	return &domain.PaymentResponse{
		Status:     domain.PaymentStatusPending,
		Provider:   domain.PaymentMethodCOD,
		GatewayRef: nil,
		PaymentURL: nil,
	}, nil
}
