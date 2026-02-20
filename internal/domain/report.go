package domain

import (
	"context"
	"errors"
	"time"
)

type ReportRepository interface {
	GetSalesReport(ctx context.Context, req *SalesReportRequest) (SalesReportResponse, error)

	GetTopSellingProducts(ctx context.Context, req *TopSellingRequest) ([]TopStatItem, error)
	GetTopSellingCategories(ctx context.Context, req *TopSellingRequest) ([]TopStatItem, error)
	GetTopSellingBrands(ctx context.Context, req *TopSellingRequest) ([]TopStatItem, error)
}

type SalesReportRequest struct {
	FromDate time.Time `json:"from"`
	ToDate   time.Time `json:"to"`
	// Total         float64   `json:"total"`
	// TotalPaid     float64   `json:"total_paid"`
	// TotalRefunded float64   `json:"total_refunded"`
}

type SalesReportResponse struct {
	SalesReportRequest
	SalesSummary
}

type SalesSummary struct {
	TotalOrders    int64   `json:"total_orders"`
	GrossRevenue   float64 `json:"gross_revenue"`
	ReturnedAmount float64 `json:"returned_amount"`
	CouponDiscount float64 `json:"coupon_discount"`
	NetRevenue     float64 `json:"net_revenue"`
}

type TopSellingRequest struct {
	Type  string    `form:"type"` // product | category | brand
	From  time.Time `form:"from" time_format:"2006-01-02"`
	To    time.Time `form:"to" time_format:"2006-01-02"`
	Limit int       `form:"limit"`
}

type TopSellingResponse struct {
	Type  string        `json:"type"` // product | category | brand
	From  time.Time     `json:"from,omitempty"`
	To    time.Time     `json:"to,omitempty"`
	Items []TopStatItem `json:"items"`
}

type TopStatItem struct {
	ID           int64   `json:"id"`
	Name         string  `json:"name"`
	TotalSold    int64   `json:"total_sold"`
	TotalRevenue float64 `json:"total_revenue"`
}

func (r *TopSellingRequest) Validate(now time.Time) error {

	validTypes := map[string]bool{
		"product":  true,
		"category": true,
		"brand":    true,
	}

	if !validTypes[r.Type] {
		return errors.New("invalid type")
	}

	// Default dates
	if r.From.IsZero() {
		r.From = now.AddDate(0, 0, -30)
	}
	if r.To.IsZero() {
		r.To = now
	}

	// Date validation
	if r.From.After(now) {
		return errors.New("from date cannot be in the future")
	}
	if r.From.After(r.To) {
		return errors.New("from date cannot be after to date")
	}
	if r.To.After(now) {
		return errors.New("to date cannot be in the future")
	}

	// Limit handling
	if r.Limit == 0 {
		r.Limit = 10
	} else if r.Limit < 0 || r.Limit > 10 {
		return errors.New("limit must be between 1 and 10")
	}

	return nil
}
