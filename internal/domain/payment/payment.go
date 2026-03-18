package payment

import (
	"context"
)

type PaymentMethod string
type PaymentStatus string

const (
	PaymentMethodRazorpay PaymentMethod = "RAZORPAY"
	PaymentMethodCOD      PaymentMethod = "COD"
	PaymentMethodWallet   PaymentMethod = "WALLET"
)

var AvailablePaymentMethods = []PaymentMethod{
	PaymentMethodRazorpay,
	PaymentMethodCOD,
	PaymentMethodWallet,
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

type Request struct {
	OrderID  string
	UserID   int64
	Amount   int64
	Currency string
}

type Response struct {
	Status     PaymentStatus
	Provider   PaymentMethod
	GatewayRef *string
	PaymentURL *string
}

type WebhookEvent struct {
	Event   string `json:"event"`
	Payload struct {
		Payment struct {
			Entity struct {
				PaymentID string `json:"id"`
				Amount    int64  `json:"amount"`
				Status    string `json:"status"`
				OrderID   string `json:"order_id"`
				Captured  bool   `json:"captured"`
				Notes     struct {
					InternalOrderID string `json:"internal_order_id"`
					UserID          string `json:"user_id"`
				} `json:"notes"`
			} `json:"entity"`
		} `json:"payment"`
	} `json:"payload"`
}

type WebhookPayment struct {
	ID        string            `json:"id"`
	OrderID   string            `json:"order_id"`
	Status    string            `json:"status"`
	Captured  bool              `json:"captured"`
	Amount    int64             `json:"amount"`
	Notes     map[string]string `json:"notes"`
	CreatedAt int64             `json:"created_at"`
}

//go:generate mockgen -destination=mocks/mock_payment.go -package=mocks . PaymentGateway
type PaymentGateway interface {
	CreatePayment(ctx context.Context, req Request) (*Response, error)
}
