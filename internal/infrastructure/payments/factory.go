package payments

import (
	"github.com/razorpay/razorpay-go"
	"github.com/tharunn0/E-Commerce-Go/internal/domain"
)

func GetPaymentGateway(method domain.PaymentMethod, client *razorpay.Client) domain.PaymentGateway {
	switch method {
	case domain.PaymentMethodRazorpay:
		return &RazorpayPayment{Client: client}
	case domain.PaymentMethodCOD:
		return &CODPayment{}
	default:
		return nil
	}
}
