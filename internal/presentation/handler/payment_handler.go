package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/tharunn0/E-Commerce-Go/internal/config"
	"github.com/tharunn0/E-Commerce-Go/internal/infrastructure/payments"
	"github.com/tharunn0/E-Commerce-Go/internal/utils"
	"go.uber.org/zap"
)

type PaymentHandler struct {
	razorpayCfg *config.RazorpaySettings
	logger      *zap.Logger
}

func NewPaymentHandler(razorpayCfg *config.RazorpaySettings, logger *zap.Logger) *PaymentHandler {
	return &PaymentHandler{
		razorpayCfg: razorpayCfg,
		logger:      logger,
	}
}

func (h *PaymentHandler) SimulatePayment(c *gin.Context) {
	type Request struct {
		OrderID string `json:"order_id" binding:"required"`
		Status  string `json:"status"` // success or failed
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	// Default = success
	if req.Status == "failed" {
		c.JSON(400, gin.H{
			"error": gin.H{
				"code":        "BAD_REQUEST_ERROR",
				"description": "Payment failed",
				"reason":      "payment_failed",
			},
		})
		return
	}

	paymentID := "pay_" + utils.RandomString(14)

	signature := payments.GenerateRazorpaySignature(req.OrderID, h.razorpayCfg.KeySecret, paymentID) // uses your TEST secret

	c.JSON(200, gin.H{
		"razorpay_order_id":   req.OrderID,
		"razorpay_payment_id": paymentID,
		"razorpay_signature":  signature,
	})
}

func (h *PaymentHandler) VerifyPayment(c *gin.Context) {
	type Request struct {
		OrderID   string `json:"razorpay_order_id" binding:"required"`
		PaymentID string `json:"razorpay_payment_id" binding:"required"`
		Signature string `json:"razorpay_signature" binding:"required"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"status": "failed", "message": "invalid payload"})
		return
	}

	expected := payments.GenerateRazorpaySignature(req.OrderID, h.razorpayCfg.KeySecret, req.PaymentID)

	if expected == req.Signature {
		c.JSON(200, gin.H{
			"status":  "success",
			"message": "payment verified",
		})
	} else {
		c.JSON(400, gin.H{
			"status":  "failed",
			"message": "invalid signature",
		})
	}
}
