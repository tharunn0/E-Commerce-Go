package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tharunn0/E-Commerce-Go/internal/config"
	"github.com/tharunn0/E-Commerce-Go/internal/domain"
	"github.com/tharunn0/E-Commerce-Go/internal/infrastructure/payments"
	"github.com/tharunn0/E-Commerce-Go/internal/service"

	"github.com/razorpay/razorpay-go"
	"go.uber.org/zap"
)

type PaymentHandler struct {
	orderService   *service.OrderService
	paymentService *service.PaymentService
	razorpayClient *razorpay.Client
	razorpayCfg    config.RazorpaySettings
	logger         *zap.Logger
}

func NewPaymentHandler(orderService *service.OrderService, paymentService *service.PaymentService, razorpayCfg config.RazorpaySettings, logger *zap.Logger, razorpayClient *razorpay.Client) *PaymentHandler {
	return &PaymentHandler{
		orderService:   orderService,
		paymentService: paymentService,
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	res, apierr := h.paymentService.CreatePaymentLink(ctx, req.InternalOrderID)
	if apierr != nil {
		c.JSON(apierr.Status, gin.H{"error": apierr.Code, "message": apierr.Message})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": res})
}

func (h *PaymentHandler) Webhook(c *gin.Context) {

	ctx := c.Request.Context()

	h.logger.Info("webhook request received")

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Error("failed to read request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	signature := c.GetHeader("X-Razorpay-Signature")

	if !payments.VerifyRazorpaySignature(body, signature, h.razorpayCfg.WebhookSecret) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "INVALID_AUTHORIZATION_HEADER"})
		return
	}

	var req domain.WebhookEvent

	if err := json.Unmarshal(body, &req); err != nil {
		h.logger.Error("failed to unmarshal request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// h.logger.Info("Body :",
	// 	zap.String("event", req.Event),
	// 	zap.String("status", req.Payload.Payment.Entity.Status),
	// 	zap.Bool("captured", req.Payload.Payment.Entity.Captured),
	// 	zap.Int64("amount", req.Payload.Payment.Entity.Amount),
	// 	zap.String("order_id", req.Payload.Payment.Entity.OrderID),
	// 	zap.String("internal_order_id", req.Payload.Payment.Entity.Notes.InternalOrderID),
	// 	zap.String("user_id", req.Payload.Payment.Entity.Notes.UserID),
	// 	zap.String("payment_id", req.Payload.Payment.Entity.PaymentID),
	// )

	apierr := h.paymentService.VerifyPayment(ctx, &req)
	if apierr != nil {
		c.JSON(apierr.Status, gin.H{"error": apierr.Code, "message": apierr.Message})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "webhook request processed successfully"})

}
