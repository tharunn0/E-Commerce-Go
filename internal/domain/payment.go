package domain

import (
	"context"
)

type PaymentRepository interface {
	CreatePayment(ctx context.Context, payment *Payment) error
}

type PaymentMethod string
type PaymentStatus string

const (
	PaymentMethodRazorpay PaymentMethod = "RAZORPAY"
	PaymentMethodCOD      PaymentMethod = "COD"
)

var AvailablePaymentMethods = []PaymentMethod{
	PaymentMethodRazorpay,
	PaymentMethodCOD,
}

const (
	PaymentStatusPending   PaymentStatus = "PENDING"
	PaymentStatusCompleted PaymentStatus = "COMPLETED"
	PaymentStatusFailed    PaymentStatus = "FAILED"
	PaymentStatusRefunded  PaymentStatus = "REFUNDED"
)

type Payment struct {
	OrderID    string        `json:"order_id"`
	UserID     int64         `json:"user_id"`
	Amount     int64         `json:"amount"`
	Currency   string        `json:"currency"`
	Status     PaymentStatus `json:"status"`
	Provider   PaymentMethod `json:"provider"`
	GatewayRef *string       `json:"gateway_ref,omitempty"`
	PaymentURL *string       `json:"payment_url,omitempty"`
}

type PaymentRequest struct {
	OrderID  string
	UserID   int64
	Amount   int64
	Currency string
}

type PaymentResponse struct {
	Status     PaymentStatus
	Provider   PaymentMethod
	GatewayRef *string
	PaymentURL *string
}

type PaymentGateway interface {
	CreatePayment(ctx context.Context, req PaymentRequest) (*PaymentResponse, error)
}
