package service

import (
	"context"
	"net/http"

	"github.com/tharunn0/E-Commerce-Go/internal/apperror"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/wishlist"
	"github.com/tharunn0/E-Commerce-Go/internal/utils"
	"go.uber.org/zap"
)

type WishlistService struct {
	repo wishlist.WishlistRepository
	log  *zap.Logger
}

func NewWishlistService(wishlistRepository wishlist.WishlistRepository, logger *zap.Logger) *WishlistService {
	return &WishlistService{
		repo: wishlistRepository,
		log:  logger,
	}
}

func (s *WishlistService) AddToWishlist(ctx context.Context, productID int64) (*wishlist.Wishlist, *apperror.APIError) {

	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "INVALID_USER",
			Message: "Failed to retreive user id from context",
		}
	}
	if userID == 0 {
		return nil, &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_USER",
			Message: "Invalid user ID",
		}
	}

	if productID <= 0 {
		return nil, &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_PRODUCT",
			Message: "Invalid product ID",
		}
	}

	wishlistId, err := s.repo.AddToWishlist(ctx, userID, productID)
	if err != nil {
		if err == apperror.ErrWishlistItemAlreadyExists {
			return nil, &apperror.APIError{
				Status:  http.StatusConflict,
				Code:    "WISHLIST_ALREADY_EXISTS",
				Message: "Product is already in wishlist",
			}
		}
		s.log.Error("Failed to add to wishlist", zap.String("service-func", "WishlistService.AddToWishlist"), zap.Error(err))
		return nil, &apperror.APIError{
			Status:  http.StatusNotFound,
			Code:    "INVALID_PRODUCT_ID",
			Message: "Item not found",
		}
	}

	if wishlistId == 0 {
		return nil, &apperror.APIError{
			Status:  http.StatusNotFound,
			Code:    "WISHLIST_ERROR",
			Message: "Failed to get wishlist ID",
		}
	}

	wish, err := s.repo.GetWishlistByID(ctx, wishlistId)
	if err != nil {
		s.log.Error("Failed to retrieve wishlist", zap.String("service-func", "WishlistService.AddToWishlist"), zap.Error(err))
		return nil, &apperror.APIError{
			Status:  http.StatusNotFound,
			Code:    "WISHLIST_ERROR",
			Message: "Failed to retrieve wishlist",
		}
	}
	return wish, nil
}

func (s *WishlistService) GetWishlistByUserID(ctx context.Context) (*wishlist.Wishlist, *apperror.APIError) {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, &apperror.APIError{
			Code:    "INVALID_USER",
			Message: "Failed to retrieve user id from context",
		}
	}
	if userID == 0 {
		return nil, &apperror.APIError{
			Code:    "INVALID_USER",
			Message: "Invalid user ID",
		}
	}

	wish, err := s.repo.GetWishlistByUserID(ctx, userID)
	if err != nil {
		return nil, &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to retrieve wishlist",
		}
	}
	return wish, nil
}

func (s *WishlistService) RemoveFromWishlist(ctx context.Context, productIDs []int64) *apperror.APIError {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return &apperror.APIError{
			Code:    "INVALID_USER",
			Message: "Failed to retrieve user id from context",
		}
	}
	if userID == 0 {
		return &apperror.APIError{
			Code:    "INVALID_USER",
			Message: "Invalid user ID",
		}
	}

	if len(productIDs) == 0 {
		return &apperror.APIError{
			Code:    "INVALID_INPUT",
			Message: "No product IDs provided",
		}
	}

	err = s.repo.RemoveFromWishlist(ctx, userID, productIDs)
	if err != nil {
		if err == apperror.ErrWishlistItemNotFound {
			return &apperror.APIError{
				Status:  404,
				Code:    "WISHLIST_ITEM_NOT_FOUND",
				Message: "Wishlist item not found",
			}
		}
		return &apperror.APIError{
			Status:  500,
			Code:    "DB_ERROR",
			Message: "Failed to remove from wishlist",
		}
	}

	return nil
}
