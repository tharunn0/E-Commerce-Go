package domain

import "context"

type PaymentMethod string
type PaymentStatus string

const (
	PaymentMethodRazorpay PaymentMethod = "RAZORPAY"
	PaymentMethodCOD      PaymentMethod = "COD"
)

const (
	PaymentStatusPending   PaymentStatus = "PENDING"
	PaymentStatusCompleted PaymentStatus = "COMPLETED"
	PaymentStatusFailed    PaymentStatus = "FAILED"
	PaymentStatusRefunded  PaymentStatus = "REFUNDED"
)

type PaymentRequest struct {
	OrderID  string
	UserID   int64
	Amount   int64
	Currency string
}

type PaymentResponse struct {
	Status     PaymentStatus
	GatewayRef string
	PaymentURL string
}

type PaymentGateway interface {
	CreatePayment(ctx context.Context, req PaymentRequest) (*PaymentResponse, error)
}
