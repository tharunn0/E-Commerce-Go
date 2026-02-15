package service

import (
	"context"
	"net/http"

	"github.com/tharunn0/E-Commerce-Go/internal/apperror"
	"github.com/tharunn0/E-Commerce-Go/internal/domain"
	"go.uber.org/zap"
)

type CouponService struct {
	repo domain.CouponRepository
	log  *zap.Logger
}

func NewCouponService(repo domain.CouponRepository, log *zap.Logger) *CouponService {
	return &CouponService{repo: repo, log: log}
}

func (serv *CouponService) CreateCoupon(ctx context.Context, req *domain.CreateCouponRequest) (*domain.CouponResponse, *apperror.APIError) {

	// validate coupon request
	if err := req.Validate(); err != nil {
		return nil, &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_COUPON_REQUEST",
			Message: err.Error(),
		}
	}

	// check if coupon code already exists
	created, err := serv.repo.CreateCoupon(ctx, req)
	if err != nil {

		serv.log.Error("coupon creation failed", zap.Error(err), zap.String("coupon_code", req.CouponCode))

		if err == apperror.ErrCouponCodeAlreadyExists {
			return nil, &apperror.APIError{
				Status:  http.StatusConflict,
				Code:    "COUPON_CODE_ALREADY_EXISTS",
				Message: err.Error(),
			}
		}

		return nil, &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "COUPON_CREATION_FAILED",
			Message: "Failed to create coupon",
		}
	}

	return created, nil
}
