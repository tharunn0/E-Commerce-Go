package payments

import (
	"github.com/razorpay/razorpay-go"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/payment"
)

func GetPaymentGateway(method payment.PaymentMethod, client *razorpay.Client) payment.PaymentGateway {
	switch method {
	case payment.PaymentMethodRazorpay:
		return &RazorpayPayment{}
	case payment.PaymentMethodCOD:
		return &CODPayment{}
	case payment.PaymentMethodWallet:
		return &WalletPayment{}
	default:
		return nil
	}
}
