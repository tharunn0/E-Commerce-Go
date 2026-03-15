package payments

import (
	"context"

	"github.com/tharunn0/E-Commerce-Go/internal/domain"
)

type WalletPayment struct{}

func (w *WalletPayment) CreatePayment(ctx context.Context, req domain.PaymentRequest) (*domain.PaymentResponse, error) {
	return &domain.PaymentResponse{
		Status:     domain.PaymentStatusPending,
		Provider:   domain.PaymentMethodWallet,
		GatewayRef: nil,
		PaymentURL: nil,
	}, nil
}
