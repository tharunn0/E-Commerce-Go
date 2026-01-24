package domain

import (
	"context"
	"time"
)

type OfferRepository interface {
	CreateOffer(ctx context.Context, req *CreateOfferRequest) error
}

type Offer struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name" binding:"required"`
	Description   string    `json:"description"`
	DiscountType  string    `json:"discount_type" binding:"required,oneof=fixed percentage"`
	DiscountValue float64   `json:"discount_value" binding:"required,gte=0"`
	StartDate     time.Time `json:"start_date" binding:"required"`
	EndDate       time.Time `json:"end_date" binding:"required"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
}

type ProductOffer struct {
	ID        int64 `json:"id"`
	OfferID   int64 `json:"offer_id" binding:"required"`
	ProductID int64 `json:"product_id" binding:"required"`
}

type CategoryOffer struct {
	ID         int64 `json:"id"`
	OfferID    int64 `json:"offer_id" binding:"required"`
	CategoryID int64 `json:"category_id" binding:"required"`
}

type CreateOfferRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`

	DiscountType  string    `json:"discount_type" binding:"required,oneof=fixed percentage"`
	DiscountValue float64   `json:"discount_value" binding:"required,gte=0"`
	StartDate     time.Time `json:"start_date" binding:"required"`
	EndDate       time.Time `json:"end_date" binding:"required"`

	Scope    string  `json:"scope"`
	ScopeIDs []int64 `json:"scope_ids" binding:"required"`
}
