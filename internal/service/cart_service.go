package service

import (
	"context"

	"github.com/tharunn0/E-Commerce-Go/internal/apperror"
	"github.com/tharunn0/E-Commerce-Go/internal/domain"
	"github.com/tharunn0/E-Commerce-Go/internal/utils"
	"go.uber.org/zap"
)

type CartService struct {
	productRepo domain.ProductRepository
	repo        domain.CartRepository
	log         *zap.Logger
}

func NewCartService(repo domain.CartRepository, productRepo domain.ProductRepository, log *zap.Logger) *CartService {
	return &CartService{repo: repo, productRepo: productRepo, log: log}
}

func (s *CartService) AddToCart(ctx context.Context, req *domain.AddToCartRequest) (*domain.Cart, *apperror.APIError) {

	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, &apperror.APIError{
			Code:    "INVALID_USER_ID",
			Message: "Invalid user ID.",
		}
	}

	if req.Quantity <= 0 {
		req.Quantity = 1
	}

	cartID, err := s.repo.AddToCart(ctx, userID, req.ProductVariantID, req.Quantity)
	if err != nil {
		if err == apperror.ErrProductVariantNotFound {
			return nil, &apperror.APIError{
				Code:    "NOT_FOUND",
				Message: apperror.ErrProductVariantNotFound.Error(),
			}
		}
		s.log.Error("failed to add to cart", zap.String("function", "AddToCart"), zap.Int64("user_id", userID), zap.Error(err))
		return nil, &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to add to cart.",
		}
	}

	cart, err := s.repo.GetCartByCartID(ctx, *cartID)
	if err != nil {

		s.log.Error("failed to get cart", zap.String("function", "GetCartByCartID"), zap.Int64("cart_id", *cartID), zap.Error(err))
		return nil, &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to get cart.",
		}
	}
	return cart, nil
}

func (s *CartService) GetCart(ctx context.Context) (*domain.Cart, *apperror.APIError) {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, &apperror.APIError{
			Code:    "INVALID_USER_ID",
			Message: "Invalid user ID.",
		}
	}
	cart, err := s.repo.GetCartByUserID(ctx, userID)
	if err != nil {
		s.log.Error("failed to get cart", zap.String("function", "GetCartByCartID"), zap.Int64("user_id", userID), zap.Error(err))
		return nil, &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to get cart.",
		}
	}

	return cart, nil
}

func (s *CartService) UpdateCartItemQuantity(ctx context.Context, req *domain.UpdateCartItemQuantityRequest) (*domain.Cart, *apperror.APIError) {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, &apperror.APIError{
			Code:    "INVALID_USER_ID",
			Message: "Invalid user ID.",
		}
	}

	if req.Quantity <= 0 {
		return nil, &apperror.APIError{
			Code:    "INVALID_QUANTITY",
			Message: "Quantity must be greater than 0.",
		}
	}

	updatedCartID, err := s.repo.UpdateCartItemQuantity(ctx, userID, req)
	if err != nil {
		if err == apperror.ErrCartItemNotFound {
			return nil, &apperror.APIError{
				Code:    "NOT_FOUND",
				Message: apperror.ErrCartItemNotFound.Error(),
			}
		}
		s.log.Error("failed to update cart item quantity", zap.String("function", "UpdateCartItemQuantity"), zap.Int64("user_id", userID), zap.Error(err))
		return nil, &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to update cart item quantity.",
		}
	}
	cart, err := s.repo.GetCartByCartID(ctx, updatedCartID)
	if err != nil {
		s.log.Error("failed to get cart", zap.String("function", "GetCartByCartID"), zap.Int64("cart_id", updatedCartID), zap.Error(err))
		return nil, &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to get cart.",
		}
	}
	return cart, nil
}

func (s *CartService) RemoveCartItem(ctx context.Context, req *domain.RemoveCartItemRequest) (*domain.Cart, *apperror.APIError) {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, &apperror.APIError{
			Code:    "INVALID_USER_ID",
			Message: "Invalid user ID.",
		}
	}

	removedCartID, err := s.repo.RemoveCartItem(ctx, userID, req.ProductVariantID)
	if err != nil {
		if err == apperror.ErrCartItemNotFound {
			return nil, &apperror.APIError{
				Code:    "NOT_FOUND",
				Message: apperror.ErrCartItemNotFound.Error(),
			}
		}
		s.log.Error("failed to remove cart item", zap.String("function", "RemoveCartItem"), zap.Int64("user_id", userID), zap.Error(err))
		return nil, &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to remove cart item.",
		}
	}
	cart, err := s.repo.GetCartByCartID(ctx, removedCartID)
	if err != nil {
		s.log.Error("failed to get cart", zap.String("function", "GetCartByCartID"), zap.Int64("cart_id", removedCartID), zap.Error(err))
		return nil, &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to get cart.",
		}
	}

	return cart, nil
}

func (s *CartService) EmptyCart(ctx context.Context) *apperror.APIError {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return &apperror.APIError{
			Code:    "INVALID_USER_ID",
			Message: "Invalid user ID.",
		}
	}
	err = s.repo.EmptyCart(ctx, userID)
	if err != nil {
		s.log.Error("failed to empty cart", zap.String("function", "EmptyCart"), zap.Int64("user_id", userID), zap.Error(err))
		return &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to empty cart.",
		}
	}
	return nil
}
