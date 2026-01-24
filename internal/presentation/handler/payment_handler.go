package handler

import (
	"fmt"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/tharunn0/E-Commerce-Go/internal/config"
	"github.com/tharunn0/E-Commerce-Go/internal/domain"
	"github.com/tharunn0/E-Commerce-Go/internal/infrastructure/payments"
	"github.com/tharunn0/E-Commerce-Go/internal/service"
	"github.com/tharunn0/E-Commerce-Go/internal/utils"

	"github.com/razorpay/razorpay-go"
	"go.uber.org/zap"
)

type PaymentHandler struct {
	orderService   *service.OrderService
	razorpayClient *razorpay.Client
	razorpayCfg    config.RazorpaySettings
	logger         *zap.Logger
}

func NewPaymentHandler(orderService *service.OrderService, razorpayCfg config.RazorpaySettings, logger *zap.Logger, razorpayClient *razorpay.Client) *PaymentHandler {
	return &PaymentHandler{
		orderService:   orderService,
		razorpayCfg:    razorpayCfg,
		logger:         logger,
		razorpayClient: razorpayClient,
	}
}

func (h *PaymentHandler) CreatePaymentLink(c *gin.Context) {

	ctx := c.Request.Context()

	type Request struct {
		InternalOrderID string `json:"internal_order_id"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	order, apierr := h.orderService.GetOrderByID(ctx, req.InternalOrderID)
	if apierr != nil {
		c.JSON(400, gin.H{"error": "invalid order"})
		return
	}

	data := map[string]any{
		"amount":   order.TotalAmount * 100,
		"currency": "INR",
	}

	res, err := h.razorpayClient.PaymentLink.Create(data, nil)
	if err != nil {
		h.logger.Error("failed to create payment link", zap.Error(err))
		c.JSON(400, gin.H{"error": "failed to create payment link"})
		return
	}

	fmt.Println(res)

	c.JSON(200, gin.H{"msg": res})
}

func (h *PaymentHandler) Webhook(c *gin.Context) {

	h.logger.Info("webhook request received")

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Error("failed to read request body", zap.Error(err))
		c.JSON(400, gin.H{"error": "failed to read request body"})
		return
	}

	h.logger.Info("request body", zap.String("body", string(body)))
}

func (h *PaymentHandler) SimulatePayment(c *gin.Context) {
	type Request struct {
		OrderID string `json:"order_id" binding:"required"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	_, err := h.razorpayClient.Order.Fetch(req.OrderID, nil, nil)
	if err != nil {
		c.JSON(400, gin.H{
			"error":   "BAD_REQUEST_ERROR",
			"message": "Invalid or unknown Razorpay order_id",
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

	ctx := c.Request.Context()

	type Request struct {
		OrderID   string `json:"razorpay_order_id" binding:"required"`
		PaymentID string `json:"razorpay_payment_id" binding:"required"`
		Signature string `json:"razorpay_signature" binding:"required"`

		Status string `json:"status" binding:"required"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"status": "failed", "message": "invalid payload"})
		return
	}

	expected := payments.GenerateRazorpaySignature(req.OrderID, h.razorpayCfg.KeySecret, req.PaymentID)

	var status domain.PaymentStatus
	if req.Status == "paid" && expected == req.Signature {
		status = domain.PaymentStatusCompleted
	} else {
		status = domain.PaymentStatusFailed
	}

	err := h.orderService.UpdateOrderStatusOnPayment(ctx, status, req.OrderID)
	if err != nil {
		c.JSON(400, gin.H{
			"status":  "failed",
			"message": "invalid signature",
		})
		return
	}

	c.JSON(200, gin.H{
		"status":  "success",
		"message": "status updated successfully",
	})
}
