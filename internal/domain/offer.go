package domain

import (
	"context"
	"time"
)

type OfferRepository interface {
	CreateOffer(ctx context.Context, req *CreateOfferRequest) error

	GetAllCategoryOffers(ctx context.Context) ([]*Offer, error)
	GetAllProductOffers(ctx context.Context) ([]*Offer, error)
	GetCategoryOffers(ctx context.Context, ids []int64) ([]*Offer, error)
	GetProductOffers(ctx context.Context, ids []int64) ([]*Offer, error)
	// GetOfferByID(ctx context.Context, id int64) (*Offer, error)
	// GetAllOffers(ctx context.Context) ([]Offer, error)
}

type Offer struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	DiscountType  string    `json:"discount_type"`
	DiscountValue float64   `json:"discount_value"`
	StartDate     time.Time `json:"start_date"`
	EndDate       time.Time `json:"end_date"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`

	Scope       string  `json:"scope"` // product or category
	EligibleIds []int64 `json:"eligible_ids"`
}

type ProductOffer struct {
	ID        int64 `json:"id"`
	OfferID   int64 `json:"offer_id"`
	ProductID int64 `json:"product_id"`
}

type CategoryOffer struct {
	ID         int64 `json:"id"`
	OfferID    int64 `json:"offer_id"`
	CategoryID int64 `json:"category_id"`
}

type CreateOfferRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`

	DiscountType  string    `json:"discount_type"`
	DiscountValue float64   `json:"discount_value"`
	StartDate     time.Time `json:"start_date"`
	EndDate       time.Time `json:"end_date"`

	Scope    string  `json:"scope"`
	ScopeIDs []int64 `json:"scope_ids"`
}
