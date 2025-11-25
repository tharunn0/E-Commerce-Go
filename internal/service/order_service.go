package service

import (
	"context"
	"fmt"
	"net/http"

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

func (s *OrderService) CheckoutCart(ctx context.Context) (*domain.Cart, []domain.NotEnoughStockError, *apperror.APIError) {
	// get user id from context
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, nil, &apperror.APIError{
			Status:  http.StatusUnauthorized,
			Code:    "UNAUTHORIZED",
			Message: "Unauthorized.",
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

	return cart, nil, nil
}

func (s *OrderService) CheckoutProductVariant(ctx context.Context, variantID int64, quantity int64) (*domain.ProductVariantResponse, *apperror.APIError) {
	activeonly := !utils.IsAdmin(ctx)
	productVariant, err := s.productRepo.GetProductVariantByID(ctx, variantID, activeonly)
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

	return productVariant, nil
}
