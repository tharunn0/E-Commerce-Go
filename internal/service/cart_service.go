package service

import (
	"context"
	"log"
	"net/http"

	"github.com/tharunn0/E-Commerce-Go/internal/apperror"
	"github.com/tharunn0/E-Commerce-Go/internal/domain"
	"github.com/tharunn0/E-Commerce-Go/internal/utils"
	"go.uber.org/zap"
)

type CartService struct {
	productRepo domain.ProductRepository
	offerRepo   domain.OfferRepository
	cartRepo    domain.CartRepository
	log         *zap.Logger
}

func NewCartService(
	cartRepo domain.CartRepository,
	productRepo domain.ProductRepository,
	offerRepo domain.OfferRepository,
	log *zap.Logger) *CartService {
	return &CartService{
		cartRepo:    cartRepo,
		productRepo: productRepo,
		offerRepo:   offerRepo,
		log:         log,
	}
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
	} else if req.Quantity > 5 {
		return nil, &apperror.APIError{
			Status:  http.StatusUnprocessableEntity,
			Code:    "QUANTITY_EXCEEDED",
			Message: "Quantity must be less than or equal to 5.",
		}
	}

	cartID, err := s.cartRepo.AddToCart(ctx, userID, req.ProductVariantID, req.Quantity)
	if err != nil {
		if err == apperror.ErrProductVariantNotFound {
			return nil, &apperror.APIError{
				Code:    "NOT_FOUND",
				Message: apperror.ErrProductVariantNotFound.Error(),
			}
		}
		if err == apperror.ErrQuantityExceeded {
			return nil, &apperror.APIError{
				Status:  http.StatusUnprocessableEntity,
				Code:    "QUANTITY_EXCEEDED",
				Message: apperror.ErrQuantityExceeded.Error(),
			}
		}
		if err == apperror.ErrNotEnoughStock {
			return nil, &apperror.APIError{
				Status:  http.StatusConflict,
				Code:    "NOT_ENOUGH_STOCK",
				Message: "Not enough stock.",
			}
		}
		s.log.Error("failed to add to cart", zap.String("function", "AddToCart"), zap.Int64("user_id", userID), zap.Error(err))
		return nil, &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "DB_ERROR",
			Message: "Failed to add to cart.",
		}
	}

	cart, err := s.cartRepo.GetCartByID(ctx, *cartID)
	if err != nil {

		s.log.Error("failed to get cart", zap.String("function", "GetCartByID"), zap.Int64("cart_id", *cartID), zap.Error(err))
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
	cart, err := s.cartRepo.GetCartByUserID(ctx, userID)
	if err != nil {
		s.log.Error("failed to get cart", zap.String("function", "GetCartByCartID"), zap.Int64("user_id", userID), zap.Error(err))
		return nil, &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to get cart.",
		}
	}

	if cart.Items == nil {
		return cart, nil
	}

	variantInfo, err := s.cartRepo.GetCartVariantInfo(ctx, userID)
	if err != nil {
		s.log.Error("failed to get variant stocks", zap.String("function", "GetCartVariantInfo"), zap.Error(err))
		return nil, &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to get variant stocks.",
		}
	}

	var productIDs []int64
	var categoryIDs []int64
	for _, item := range cart.Items {
		productIDs = append(productIDs, item.ProductID)
		categoryIDs = append(categoryIDs, item.CategoryID)
		if variantInfo[item.ProductVariantID].Stock < item.Quantity {
			if variantInfo[item.ProductVariantID].Stock == 0 {
				item.Status = "OUT_OF_STOCK"
			} else {
				item.Status = "LOW_STOCK"
			}
			item.Message = "Quantity is greater than stock."
		}
	}

	offers, err := s.offerRepo.GetAllActiveOffers(ctx, productIDs, categoryIDs)
	if err != nil {
		s.log.Error("failed to get offers", zap.String("function", "GetAllActiveOffers"), zap.Error(err))
		return nil, &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to get offers.",
		}
	}

	if len(offers) == 0 {
		log.Println("no offers found", "function", "GetAllActiveOffers")
	} else {
		log.Println("offers found", "function", "GetAllActiveOffers")
	}

	domain.ApplyDiscounts(cart, offers)
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

	updatedCartID, err := s.cartRepo.UpdateCartItemQuantity(ctx, userID, req)
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
	cart, err := s.cartRepo.GetCartByID(ctx, updatedCartID)
	if err != nil {
		s.log.Error("failed to get cart", zap.String("function", "GetCartByID"), zap.Int64("cart_id", updatedCartID), zap.Error(err))
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

	removedCartID, err := s.cartRepo.RemoveCartItem(ctx, userID, req.ProductVariantID)
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
	cart, err := s.cartRepo.GetCartByID(ctx, removedCartID)
	if err != nil {
		s.log.Error("failed to get cart", zap.String("function", "GetCartByID"), zap.Int64("cart_id", removedCartID), zap.Error(err))
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
	err = s.cartRepo.EmptyCart(ctx, userID)
	if err != nil {
		s.log.Error("failed to empty cart", zap.String("function", "EmptyCart"), zap.Int64("user_id", userID), zap.Error(err))
		return &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to empty cart.",
		}
	}
	return nil
}
