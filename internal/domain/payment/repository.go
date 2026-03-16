package payment

import "context"

type PaymentRepository interface {
	CreatePayment(ctx context.Context, payment *Payment) error
	UpdatePaymentStatus(ctx context.Context, orderID string, status PaymentStatus) error
	UpdateOrderPaymentStatus(ctx context.Context, orderID string, paymentData *WebhookPayment) error
}
