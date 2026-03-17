package service

import (
	"context"
	"errors"
	"net/http"

	"github.com/tharunn0/E-Commerce-Go/internal/apperror"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/review"
	"github.com/tharunn0/E-Commerce-Go/internal/utils"
	"go.uber.org/zap"
)

type ReviewService struct {
	repo review.ReviewRepository
	log  *zap.Logger
}

func NewReviewService(repo review.ReviewRepository, log *zap.Logger) *ReviewService {
	return &ReviewService{repo: repo, log: log}
}

func (s *ReviewService) CreateReview(ctx context.Context, req *review.CreateReviewRequest) *apperror.APIError {
	// extract user id from context
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		s.log.Error("Failed to get user id", zap.Error(err))
		return apperror.New(http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user id")
	}
	req.UserID = userID

	if err := req.Validate(); err != nil {
		s.log.Error("[service:review_service:CreateReview]", zap.String("msg", "Invalid review"), zap.Error(err))
		return apperror.New(http.StatusBadRequest, "INVALID_REVIEW", err.Error())
	}

	err = s.repo.CreateReview(ctx, req)
	if err != nil {
		s.log.Error("[service:review_service:CreateReview]", zap.String("msg", "Failed to create review"), zap.Error(err))
		if errors.Is(err, apperror.ErrReviewAlreadyExists) {
			return apperror.New(http.StatusConflict, "REVIEW_ALREADY_EXISTS", err.Error())
		}
		if errors.Is(err, apperror.ErrFailedToCreateReview) {
			return apperror.New(http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Failed to create review")
		}
		if errors.Is(err, apperror.ErrProductNotOrdered) {
			return apperror.New(http.StatusBadRequest, "PRODUCT_NOT_ORDERED", "You can only review products that you have ordered")
		}
		return apperror.New(http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Failed to create review")
	}
	return nil
}

func (s *ReviewService) DeleteReview(ctx context.Context, id int64) error {
	return s.repo.DeleteReview(ctx, id)
}

func (s *ReviewService) GetProductReviews(ctx context.Context, productID int64, filter *review.ReviewFilter) (*review.ProductReviews, *apperror.APIError) {

	if err := filter.Validate(); err != nil {
		s.log.Error("[service:review_service:GetProductReviews]", zap.String("msg", "Invalid filter"), zap.Error(err))
		return nil, apperror.New(http.StatusBadRequest, "INVALID_FILTER", err.Error())
	}

	productReviews, err := s.repo.GetProductReviews(ctx, productID, filter)
	if err != nil {
		s.log.Error("[service:review_service:GetProductReviews]", zap.String("msg", "Failed to get product reviews"), zap.Error(err))
		return nil, apperror.New(http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Failed to get product reviews")
	}
	return productReviews, nil
}
