package service

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/tharunn0/E-Commerce-Go/internal/apperror"
	"github.com/tharunn0/E-Commerce-Go/internal/domain"
	"github.com/tharunn0/E-Commerce-Go/internal/utils"
	"go.uber.org/zap"
)

type CouponService struct {
	log       *zap.Logger
	repo      domain.CouponRepository
	userRepo  domain.UserRepository
	cartRepo  domain.CartRepository
	offerRepo domain.OfferRepository
}

func NewCouponService(repo domain.CouponRepository, log *zap.Logger, userRepo domain.UserRepository, cartRepo domain.CartRepository, offerRepo domain.OfferRepository) *CouponService {
	return &CouponService{repo: repo, log: log, userRepo: userRepo, cartRepo: cartRepo, offerRepo: offerRepo}
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

func (serv *CouponService) ListAllCoupons(ctx context.Context, filter *domain.ListCouponsFilter) ([]domain.CouponResponse, *apperror.APIError) {

	if !utils.IsAdmin(ctx) {
		filter = nil
	}

	log.Println("filter is nil:", filter == nil)

	coupons, err := serv.repo.ListAllCoupons(ctx, filter)
	if err != nil {
		serv.log.Error("coupon listing failed", zap.Error(err))
		return nil, &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "COUPON_LISTING_FAILED",
			Message: "Failed to list coupons",
		}
	}

	return coupons, nil
}

func (s *CouponService) ApplyCoupon(ctx context.Context, req *domain.ApplyCouponRequest) (*domain.CartCheckoutResponse, []domain.NotEnoughStockError, *apperror.APIError) {

	log.Println("[service] apply coupon hit")

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
		s.log.Error("Failed to get user id from context", zap.Error(err))
		return nil, nil, &apperror.APIError{
			Status:  http.StatusUnauthorized,
			Code:    "UNAUTHORIZED",
			Message: "You are not authorized to perform this action.",
		}
	}

	// validate address
	userAddr, err := s.userRepo.GetUserAddressByID(ctx, req.AddressID)
	if err != nil {
		s.log.Error("Failed to get address", zap.Error(err))
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

	// validate address
	err = utils.ValidateUserAddress(userAddr, userID)
	if err != nil {
		return nil, nil, &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "BAD_REQUEST",
			Message: err.Error(),
		}
	}

	// get cart by user id
	cart, err := s.cartRepo.GetCartByUserID(ctx, userID)
	if err != nil {
		if err == apperror.ErrCartNotFound {
			s.log.Error("Failed to get cart", zap.Error(err))
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

	if cart.Items == nil {
		return nil, nil, &apperror.APIError{
			Status:  http.StatusNotFound,
			Code:    "NOT_FOUND",
			Message: "Cart is empty.",
		}
	}

	// get cart variant

	cartVariantInfo, err := s.cartRepo.GetCartVariantInfo(ctx, userID)
	if err != nil {
		if err == apperror.ErrCartNotFound {
			s.log.Error("Failed to validate cart", zap.Error(err))
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

	// validate not enough stock
	notEnoughStockError := utils.ValidateOrderItemsStock(cartVariantInfo, cart)
	if len(notEnoughStockError) > 0 {
		return nil, notEnoughStockError, nil
	}

	var productIds, categoryIds []int64

	for _, item := range cart.Items {
		productIds = append(productIds, item.ProductID)
		categoryIds = append(categoryIds, item.CategoryID)
	}

	// apply existing offers

	offers, err := s.offerRepo.GetAllActiveOffers(ctx, productIds, categoryIds)
	if err != nil {
		s.log.Error("Failed to get offers", zap.Error(err))
		return nil, nil, &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "DB_ERROR",
			Message: "Failed to get offers.",
		}
	}

	if len(offers) > 0 {
		domain.ApplyDiscounts(cart, offers)
	}

	// apply coupons

	if req.CouponCode != "" {
		coupon, err := s.repo.FetchCoupon(ctx, req.CouponCode)
		if err != nil && err != apperror.ErrCouponNotFound {
			s.log.Error("Failed to get coupon", zap.Error(err))
			return nil, nil, &apperror.APIError{
				Status:  http.StatusInternalServerError,
				Code:    "DB_ERROR",
				Message: "Failed to get coupon.",
			}
		}

		if err == apperror.ErrCouponNotFound {
			cart.CouponData = &domain.CouponData{
				CouponCode:       req.CouponCode,
				DiscountType:     "nil",
				DiscountValue:    0,
				DiscountedAmount: 0,
				Message:          "Invalid or expired coupon",
			}
		} else {

			err = domain.ValidateCoupon(coupon, time.Now())
			if err != nil {
				return nil, nil, &apperror.APIError{
					Status:  http.StatusBadRequest,
					Code:    "BAD_REQUEST",
					Message: err.Error(),
				}
			}

			err = domain.ApplyCouponToCart(cart, coupon)
			if err != nil {
				return nil, nil, &apperror.APIError{
					Status:  http.StatusBadRequest,
					Code:    "BAD_REQUEST",
					Message: err.Error(),
				}
			}
		}
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

	s.log.Info("[service] Cart checkout successful", zap.Any("cart", cart), zap.Any("address", userAddr))

	return resp, nil, nil
}
