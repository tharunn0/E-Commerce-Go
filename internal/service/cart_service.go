package service

import (
	"context"
	"log"
	"net/http"

	"github.com/tharunn0/E-Commerce-Go/internal/apperror"
	"github.com/tharunn0/E-Commerce-Go/internal/config"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/cart"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/discount"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/product"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/promotion"
	"github.com/tharunn0/E-Commerce-Go/internal/utils"
	"go.uber.org/zap"
)

type CartService struct {
	productRepo product.ProductRepository
	offerRepo   promotion.OfferRepository
	cartRepo    cart.CartRepository
	cfg         config.CartSettings
	log         *zap.Logger
}

func NewCartService(
	cartRepo cart.CartRepository,
	productRepo product.ProductRepository,
	offerRepo promotion.OfferRepository,
	cfg config.CartSettings,
	log *zap.Logger) *CartService {
	return &CartService{
		cartRepo:    cartRepo,
		productRepo: productRepo,
		offerRepo:   offerRepo,
		cfg:         cfg,
		log:         log,
	}
}

func (s *CartService) AddToCart(ctx context.Context, req *cart.AddToCartRequest) (*cart.Cart, *apperror.APIError) {

	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, &apperror.APIError{
			Code:    "INVALID_USER_ID",
			Message: "Invalid user ID.",
		}
	}

	if req.Quantity <= 0 {
		req.Quantity = 1
	} else if req.Quantity > s.cfg.MaxQuantity {
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

	cartObj, err := s.cartRepo.GetCartByID(ctx, *cartID)
	log.Println("cart :", cartObj)
	if err != nil {

		s.log.Error("failed to get cart", zap.String("function", "GetCartByID"), zap.Int64("cart_id", *cartID), zap.Error(err))
		return nil, &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to get cart.",
		}
	}

	return cartObj, nil
}

func (s *CartService) GetCart(ctx context.Context) (*cart.Cart, *apperror.APIError) {

	utils.LogCtxContent(ctx, s.log)

	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, &apperror.APIError{
			Code:    "INVALID_USER_ID",
			Message: "Invalid user ID.",
		}
	}
	cartObj, err := s.cartRepo.GetCartByUserID(ctx, userID)
	if err != nil {
		s.log.Error("failed to get cart", zap.String("function", "GetCartByCartID"), zap.Int64("user_id", userID), zap.Error(err))
		return nil, &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to get cart.",
		}
	}

	if cartObj.Items == nil {
		return cartObj, nil
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
	for _, item := range cartObj.Items {
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

	log.Println("cart before discounts ", cartObj)

	discount.ApplyDiscounts(cartObj, offers)

	log.Println("cart after discounts ", cartObj)

	return cartObj, nil
}

func (s *CartService) UpdateCartItemQuantity(ctx context.Context, req *cart.UpdateCartItemQuantityRequest) (*cart.Cart, *apperror.APIError) {
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
	cartObj, err := s.cartRepo.GetCartByID(ctx, updatedCartID)
	if err != nil {
		s.log.Error("failed to get cart", zap.String("function", "GetCartByID"), zap.Int64("cart_id", updatedCartID), zap.Error(err))
		return nil, &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to get cart.",
		}
	}

	var productIDs []int64
	var categoryIDs []int64
	for _, item := range cartObj.Items {
		productIDs = append(productIDs, item.ProductID)
		categoryIDs = append(categoryIDs, item.CategoryID)
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

	discount.ApplyDiscounts(cartObj, offers)

	return cartObj, nil
}

func (s *CartService) RemoveCartItem(ctx context.Context, req *cart.RemoveCartItemRequest) (*cart.Cart, *apperror.APIError) {
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
	cartObj, err := s.cartRepo.GetCartByID(ctx, removedCartID)
	if err != nil {
		s.log.Error("failed to get cart", zap.String("function", "GetCartByID"), zap.Int64("cart_id", removedCartID), zap.Error(err))
		return nil, &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to get cart.",
		}
	}

	return cartObj, nil
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
