package service

import (
	"context"
	"net/http"

	"github.com/tharunn0/E-Commerce-Go/internal/apperror"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/promotion"
	"go.uber.org/zap"
)

type OfferService struct {
	repo promotion.OfferRepository
	log  *zap.Logger
}

func NewOfferService(repo promotion.OfferRepository, log *zap.Logger) *OfferService {
	return &OfferService{repo: repo, log: log}
}

func (s *OfferService) CreateOffer(ctx context.Context, req *promotion.CreateOfferRequest) *apperror.APIError {

	// validate scopes

	switch req.Scope {
	case "product":
		for _, id := range req.ScopeIDs {
			if id <= 0 {
				s.log.Error("Invalid product ID for offer", zap.Int64("id", id))
				return &apperror.APIError{
					Status:  http.StatusBadRequest,
					Code:    "INVALID_PRODUCT_ID",
					Message: "Invalid product ID",
				}
			}
		}
	case "category":
		for _, id := range req.ScopeIDs {
			if id <= 0 {
				s.log.Error("Invalid category ID for offer", zap.Int64("id", id))
				return &apperror.APIError{
					Status:  http.StatusBadRequest,
					Code:    "INVALID_CATEGORY_ID",
					Message: "Invalid category ID",
				}
			}
		}
	default:
		s.log.Error("Invalid scope for offer", zap.String("scope", req.Scope))
		return &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_SCOPE",
			Message: "Invalid scope",
		}
	}

	// validate discount values, fields

	if req.DiscountType != "fixed" && req.DiscountType != "percentage" {
		return &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_DISCOUNT_TYPE",
			Message: "Invalid discount type",
		}
	}

	if req.DiscountValue <= 0 || req.DiscountValue > 100 {
		return &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_DISCOUNT_VALUE",
			Message: "Invalid discount value",
		}
	}

	if req.StartDate.IsZero() || req.EndDate.IsZero() {
		return &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_DATE",
			Message: "Invalid date",
		}
	}

	if req.StartDate.After(req.EndDate) {
		return &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_DATE",
			Message: "Invalid date",
		}
	}

	// create offer

	err := s.repo.CreateOffer(ctx, req)
	if err != nil {
		s.log.Error("Failed to create offer", zap.Error(err))
		if err == apperror.ErrProductDoesNotExist {
			return &apperror.APIError{
				Status:  http.StatusNotFound,
				Code:    "NOT_FOUND",
				Message: "Scope ids not found",
			}
		}
		return &apperror.APIError{
			Status:  http.StatusConflict,
			Code:    "DB_ERROR",
			Message: "Failed to create offer",
		}
	}

	return nil
}

func (s *OfferService) GetAllProductOffers(ctx context.Context) ([]*promotion.Offer, *apperror.APIError) {
	// offers, err := s.repo.GetAllProductOffers(ctx)
	// if err != nil {
	// 	s.log.Error("Failed to get product offers", zap.Error(err))
	// 	return nil, &apperror.APIError{
	// 		Status:  http.StatusInternalServerError,
	// 		Code:    "DB_ERROR",
	// 		Message: "Failed to get product offers",
	// 	}
	// }
	return []*promotion.Offer{}, nil
}

func (s *OfferService) GetAllCategoryOffers(ctx context.Context) ([]*promotion.Offer, *apperror.APIError) {
	// offers, err := s.repo.GetAllCategoryOffers(ctx)
	// if err != nil {
	// 	s.log.Error("Failed to get category offers", zap.Error(err))
	// 	return nil, &apperror.APIError{
	// 		Status:  http.StatusInternalServerError,
	// 		Code:    "DB_ERROR",
	// 		Message: "Failed to get category offers",
	// 	}
	// }
	return []*promotion.Offer{}, nil
}
