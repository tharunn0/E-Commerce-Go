package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/tharunn0/E-Commerce-Go/internal/apperror"
	"github.com/tharunn0/E-Commerce-Go/internal/domain"
	"github.com/tharunn0/E-Commerce-Go/internal/utils"
	"go.uber.org/zap"
)

type OrderService struct {
	userRepo    domain.UserRepository
	productRepo domain.ProductRepository
	cartRepo    domain.CartRepository
	orderRepo   domain.OrderRepository
	log         *zap.Logger
}

func NewOrderService(userRepo domain.UserRepository, productRepo domain.ProductRepository, cartRepo domain.CartRepository, orderRepo domain.OrderRepository, log *zap.Logger) *OrderService {
	return &OrderService{
		userRepo:    userRepo,
		productRepo: productRepo,
		cartRepo:    cartRepo,
		orderRepo:   orderRepo,
		log:         log,
	}
}

func (s *OrderService) CheckoutCart(ctx context.Context, req domain.CartCheckoutRequest) (*domain.CartCheckoutResponse, []domain.NotEnoughStockError, *apperror.APIError) {

	// validate delivery type
	if req.DeliveryType != domain.DeliveryTypeNormal && req.DeliveryType != domain.DeliveryTypeExpress {
		return nil, nil, &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "BAD_REQUEST",
			Message: "Invalid delivery type.",
		}
	}

	// get user id from context
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, nil, &apperror.APIError{
			Status:  http.StatusUnauthorized,
			Code:    "UNAUTHORIZED",
			Message: "You are not authorized to perform this action.",
		}
	}

	// validate address
	userAddr, err := s.userRepo.GetUserAddressByID(ctx, req.AddressID)
	if err != nil {

		if err == apperror.ErrAddressNotFoundForUser {
			return nil, nil, &apperror.APIError{
				Status:  http.StatusNotFound,
				Code:    "NOT_FOUND",
				Message: apperror.ErrAddressNotFoundForUser.Error(),
			}
		}
		s.log.Error("Failed to get address", zap.Error(err))
		return nil, nil, &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "DB_ERROR",
			Message: "Failed to get address.",
		}
	}

	if userAddr.UserID != userID {
		return nil, nil, &apperror.APIError{
			Status:  http.StatusUnauthorized,
			Code:    "UNAUTHORIZED",
			Message: "You do not have access to this address.",
		}
	}

	// get cart by user id
	cart, err := s.cartRepo.GetCartByUserID(ctx, userID)
	if err != nil {
		if err == apperror.ErrCartNotFound {
			return nil, nil, &apperror.APIError{
				Status:  http.StatusNotFound,
				Code:    "NOT_FOUND",
				Message: apperror.ErrCartNotFound.Error(),
			}
		}
		s.log.Error("Failed to get cart", zap.Error(err))
		return nil, nil, &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "DB_ERROR",
			Message: "Failed to get cart.",
		}
	}

	// get cart variant

	fmt.Println(ctx, userID)

	cartVariantStocks, err := s.cartRepo.GetCartVariantStocks(ctx, userID)
	if err != nil {
		if err == apperror.ErrCartNotFound {
			return nil, nil, &apperror.APIError{
				Status:  http.StatusNotFound,
				Code:    "NOT_FOUND",
				Message: apperror.ErrCartNotFound.Error(),
			}
		}
		s.log.Error("Failed to validate cart", zap.Error(err))
		return nil, nil, &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "DB_ERROR",
			Message: "Failed to validate cart.",
		}
	}

	var notEnoughStockErrors []domain.NotEnoughStockError

	// validate cart variant stocks
	for _, cartItem := range cart.Items {
		if cartVariantStocks[cartItem.ProductVariantID] < cartItem.Quantity {
			notEnoughStockErrors = append(notEnoughStockErrors, domain.NotEnoughStockError{
				ProductVariantID: cartItem.ProductVariantID,
				SKU:              cartItem.SKU,
				Quantity:         cartItem.Quantity,
				Stock:            cartVariantStocks[cartItem.ProductVariantID],
			})
		}
	}

	if len(notEnoughStockErrors) > 0 {
		return nil, notEnoughStockErrors, nil
	}

	// shipping charge
	shippingAmount := domain.DeliveryTypeCharges[req.DeliveryType]

	// delivery time and date
	estimatedDeliveryTime, err := domain.GetDeliveryDays(userAddr.District)
	if err != nil {
		estimatedDeliveryTime = 7
		shippingAmount = 150
	}
	estimatedDeliveryDate, err := domain.CalculateDeliveryDate(userAddr.District)
	if err != nil {
		estimatedDeliveryDate = time.Now().AddDate(0, 0, estimatedDeliveryTime)
	}

	// final cart items price
	totalAmount := cart.CartTotalPrice + shippingAmount

	resp := &domain.CartCheckoutResponse{
		Cart:                  cart,
		ShippingCost:          shippingAmount,
		TotalAmount:           totalAmount,
		DeliveryType:          req.DeliveryType,
		Address:               userAddr,
		EstimatedDeliveryTime: fmt.Sprintf("%d days", estimatedDeliveryTime),
		EstimatedDeliveryDate: estimatedDeliveryDate.String(),
	}

	return resp, nil, nil
}

func (s *OrderService) CheckoutProductVariant(ctx context.Context, req *domain.ProductVariantCheckoutRequest) (*domain.ProductVariantCheckoutResponse, *apperror.APIError) {
	activeonly := !utils.IsAdmin(ctx)

	// validate delivery type
	if req.DeliveryType != domain.DeliveryTypeNormal && req.DeliveryType != domain.DeliveryTypeExpress {
		return nil, &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "BAD_REQUEST",
			Message: "Invalid delivery type.",
		}
	}

	// get user id from context
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, &apperror.APIError{
			Status:  http.StatusUnauthorized,
			Code:    "UNAUTHORIZED",
			Message: "You are not authorized to perform this action.",
		}
	}

	// validate address exists
	userAddr, err := s.userRepo.GetUserAddressByID(ctx, req.AddressID)
	if err != nil {
		if err == apperror.ErrAddressNotFoundForUser {
			return nil, &apperror.APIError{
				Status:  http.StatusNotFound,
				Code:    "NOT_FOUND",
				Message: apperror.ErrAddressNotFoundForUser.Error(),
			}
		}
		s.log.Error("Failed to get address", zap.Error(err))
		return nil, &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "DB_ERROR",
			Message: "Failed to get address.",
		}
	}

	// validate address belongs to user
	if userAddr.UserID != userID {
		return nil, &apperror.APIError{
			Status:  http.StatusUnauthorized,
			Code:    "UNAUTHORIZED",
			Message: "You do not have access to this address.",
		}
	}

	productVariant, err := s.productRepo.GetProductVariantByID(ctx, req.ProductVariantID, activeonly)
	if err != nil {
		if err == apperror.ErrProductVariantNotFound {
			return nil, &apperror.APIError{
				Status:  http.StatusNotFound,
				Code:    "NOT_FOUND",
				Message: apperror.ErrProductVariantNotFound.Error(),
			}
		}
		s.log.Error("Failed to get product variant", zap.Error(err))
		return nil, &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "DB_ERROR",
			Message: "Failed to get product variant.",
		}
	}

	if productVariant.Stock < int(req.Quantity) {
		return nil, &apperror.APIError{
			Status:  http.StatusNotFound,
			Code:    "NOT_FOUND",
			Message: apperror.ErrNoStock.Error(),
		}
	}

	// final shipping charge
	shippingAmount := domain.DeliveryTypeCharges[req.DeliveryType]

	// final delivery time and date
	var totalAmount float64
	estimatedDeliveryTime, err := domain.GetDeliveryDays(userAddr.District)
	if err != nil {
		estimatedDeliveryTime = 7
		shippingAmount = 150
	}
	estimatedDeliveryDate, err := domain.CalculateDeliveryDate(userAddr.District)
	if err != nil {
		estimatedDeliveryDate = time.Now().AddDate(0, 0, estimatedDeliveryTime)
	}

	// final product variant price
	if productVariant.SalePrice == nil {
		totalAmount = productVariant.OriginalPrice + shippingAmount
	} else {
		totalAmount = *productVariant.SalePrice + shippingAmount
	}

	resp := &domain.ProductVariantCheckoutResponse{
		ProductVariant:        productVariant,
		ShippingCost:          shippingAmount,
		TotalAmount:           totalAmount,
		DeliveryType:          req.DeliveryType,
		Address:               userAddr,
		EstimatedDeliveryTime: fmt.Sprintf("%d days", estimatedDeliveryTime),
		EstimatedDeliveryDate: estimatedDeliveryDate.String(),
	}

	return resp, nil
}
