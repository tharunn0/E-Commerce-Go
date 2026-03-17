package promotion

import (
	"time"
)

type Offer struct {
	ID            int64
	Name          string
	Description   string
	DiscountType  string
	DiscountValue float64
	StartDate     time.Time
	EndDate       time.Time
	IsActive      bool
	CreatedAt     time.Time

	ProductIDs  []int64
	CategoryIDs []int64
}

type AppliedOfferData struct {
	OfferID        int64
	OfferName      string
	DiscountType   string
	DiscountValue  float64
	DiscountAmount float64
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
