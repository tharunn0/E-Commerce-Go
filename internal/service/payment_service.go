package service

import (
	"context"
	"net/http"

	"github.com/razorpay/razorpay-go"
	"github.com/tharunn0/E-Commerce-Go/internal/apperror"
	"github.com/tharunn0/E-Commerce-Go/internal/domain"
	"github.com/tharunn0/E-Commerce-Go/internal/utils"
	"go.uber.org/zap"
)

type PaymentService struct {
	orderRepo      domain.OrderRepository
	paymentRepo    domain.PaymentRepository
	userRepo       domain.UserRepository
	razorpayClient *razorpay.Client
	log            *zap.Logger
}

func NewPaymentService(orderRepo domain.OrderRepository, paymentRepo domain.PaymentRepository, userRepo domain.UserRepository, razorpayClient *razorpay.Client, log *zap.Logger) *PaymentService {
	return &PaymentService{
		orderRepo:      orderRepo,
		paymentRepo:    paymentRepo,
		userRepo:       userRepo,
		razorpayClient: razorpayClient,
		log:            log,
	}
}

func (s *PaymentService) CreatePaymentLink(ctx context.Context, orderID string) (map[string]any, *apperror.APIError) {

	// fetch user id from ctx
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, &apperror.APIError{
			Status:  http.StatusUnauthorized,
			Code:    "UNAUTHORIZED",
			Message: "Unauthorized",
		}
	}

	// verify order exist and fetch order details
	order, err := s.orderRepo.GetUserOrderByID(ctx, orderID, userID)
	if err != nil {
		return nil, &apperror.APIError{
			Status:  http.StatusNotFound,
			Code:    "NOT_FOUND",
			Message: "Order not found",
		}
	}

	// check if order is already confirmed and paid
	if order.Status == "confirmed" && order.PaymentStatus == "paid" {
		return nil, &apperror.APIError{
			Status:  http.StatusConflict,
			Code:    "CONFLICT",
			Message: "Order is already confirmed and paid",
		}
	}

	order = utils.CalculateOrderTotalAmount(order)

	// get user details
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, &apperror.APIError{
			Status:  http.StatusNotFound,
			Code:    "NOT_FOUND",
			Message: "User not found",
		}
	}

	// create payment link
	data := map[string]any{
		"amount":   order.TotalAmount * 100,
		"currency": "INR",
		"customer": map[string]any{
			"name":    user.FirstName + " " + user.LastName,
			"contact": user.Phone,
			"email":   user.Email,
		},
		"notes": map[string]any{
			"internal_order_id": orderID,
			"user_id":           userID,
		},
	}

	res, err := s.razorpayClient.PaymentLink.Create(data, nil)
	if err != nil {
		s.log.Warn("failed to create payment link", zap.Error(err))
		if err.Error() == "amount exceeds maximum amount allowed." {
			return nil, &apperror.APIError{
				Status:  http.StatusBadRequest,
				Code:    "BAD_REQUEST",
				Message: "Amount exceeds maximum amount allowed in current payment gateway.",
			}
		}
		return nil, &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to create payment link",
		}
	}

	return res, nil
}

func (s *PaymentService) VerifyPayment(ctx context.Context, event *domain.WebhookEvent) *apperror.APIError {

	// verify payment link

	fetchedPaymentData, err := s.razorpayClient.Payment.Fetch(event.Payload.Payment.Entity.PaymentID, nil, nil)
	if err != nil {
		s.log.Error("failed to fetch payment", zap.Error(err))
		return &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to fetch payment",
		}
	}

	fetchedPayment := domain.WebhookPayment{}

	err = utils.MapToStruct(fetchedPaymentData, &fetchedPayment)
	if err != nil {
		s.log.Error("failed to map payment", zap.Error(err))
		return &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to map payment",
		}
	}

	if fetchedPayment.Status != "captured" || !fetchedPayment.Captured {
		s.log.Error("payment not captured", zap.Error(err))
		return &apperror.APIError{
			Status:  http.StatusNotFound,
			Code:    "PAYMENT_NOT_CAPTURED",
			Message: "Payment failed. Please try again.",
		}
	}

	// update order and payment status

	err = s.paymentRepo.UpdateOrderPaymentStatus(ctx, event.Payload.Payment.Entity.Notes.InternalOrderID, &fetchedPayment)
	if err != nil {
		s.log.Error("failed to update order payment status", zap.Error(err))
		return &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to update order payment status",
		}
	}

	s.log.Info("payment verified successfully", zap.String("order_id", event.Payload.Payment.Entity.Notes.InternalOrderID))

	return nil
}
