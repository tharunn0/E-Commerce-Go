package payments

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	"github.com/razorpay/razorpay-go"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/payment"
)

type RazorpayPayment struct {
	Client *razorpay.Client
}

func (r *RazorpayPayment) CreatePayment(ctx context.Context, req payment.Request) (*payment.Response, error) {

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

	gref := rpOrder["id"].(string)
	url := "RAZORPAY_CHECKOUT_FRONTEND"

	return &payment.Response{
		Status:     payment.PaymentStatusPending,
		Provider:   payment.PaymentMethodRazorpay,
		GatewayRef: &gref,
		PaymentURL: &url,
	}, nil

}

func GenerateRazorpaySignature(orderID, razorpayTestSecret, paymentID string) string {
	mac := hmac.New(sha256.New, []byte(razorpayTestSecret))
	mac.Write([]byte(orderID + "|" + paymentID))
	return hex.EncodeToString(mac.Sum(nil))
}

func VerifyRazorpaySignature(body []byte, signature string, webhookSecret string) bool {

	mac := hmac.New(sha256.New, []byte(webhookSecret))
	mac.Write(body)

	expectedSig := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSig))
}
