package payments

import (
	"context"
	"github.com/razorpay/razorpay-go"
	"github.com/tharunn0/E-Commerce-Go/internal/domain"
)

type RazorpayPayment struct {
	Client *razorpay.Client
}

func (r *RazorpayPayment) CreatePayment(ctx context.Context, req domain.PaymentRequest) (*domain.PaymentResponse, error) {

	// create payment
	data := map[string]interface{}{
		"amount":   req.Amount * 100,
		"currency": req.Currency,
		"receipt":  req.OrderID,
	}

	rpOrder, err := r.Client.Order.Create(data, nil)
	if err != nil {
		return nil, err
	}

	return &domain.PaymentResponse{
		Status:     domain.PaymentStatusPending,
		GatewayRef: rpOrder["id"].(string),
		PaymentURL: "RAZORPAY_CHECKOUT_FRONTEND",
	}, nil

}
