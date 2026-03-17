package payments

import (
	"context"

	"github.com/tharunn0/E-Commerce-Go/internal/domain/payment"
)

type WalletPayment struct{}

func (w *WalletPayment) CreatePayment(ctx context.Context, req payment.Request) (*payment.Response, error) {
	return &payment.Response{
		Status:     payment.PaymentStatusPending,
		Provider:   payment.PaymentMethodWallet,
		GatewayRef: nil,
		PaymentURL: nil,
	}, nil
}
