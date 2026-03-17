package promotion

import (
	"context"
)

type OfferRepository interface {
	CreateOffer(ctx context.Context, req *CreateOfferRequest) error

	// GetAllCategoryOffers(ctx context.Context) ([]*Offer, error)
	// GetAllProductOffers(ctx context.Context) ([]*Offer, error)
	// GetCategoryOffers(ctx context.Context, ids []int64) ([]*Offer, error)
	// GetProductOffers(ctx context.Context, ids []int64) ([]*Offer, error)

	GetAllActiveOffers(ctx context.Context, productIDs []int64, categoryIDs []int64) ([]*Offer, error)

	// GetOfferByID(ctx context.Context, id int64) (*Offer, error)
	// GetAllOffers(ctx context.Context) ([]Offer, error)
}

type CouponRepository interface {
	CreateCoupon(ctx context.Context, req *CreateCouponRequest) (*CouponResponse, error)

	ListAllCoupons(ctx context.Context, filter *ListCouponsFilter) ([]CouponResponse, error)

	ListReferralRewards(ctx context.Context, userID int64) ([]CouponResponse, error)

	FetchCoupon(ctx context.Context, couponCode string, userID int64) (*CouponResponse, error)
}
