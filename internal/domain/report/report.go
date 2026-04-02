package report

import (
	"errors"
	"log"
	"time"
)

type SalesReportRequest struct {
	From time.Time `form:"from" time_format:"2006-01-02"`
	To   time.Time `form:"to" time_format:"2006-01-02"`
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

type RevenueAnalyticsRequest struct {
	From     time.Time `form:"from" time_format:"2006-01-02"`
	To       time.Time `form:"to" time_format:"2006-01-02"`
	Interval string    `form:"interval"`
}

type RevenueAnalyticsResponse struct {
	RevenueAnalyticsRequest
	RevenueData []RevenueData `json:"revenue_data"`
}

type RevenueData struct {
	Date    time.Time `json:"date"`
	Revenue float64   `json:"revenue"`
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

func (r *SalesReportRequest) Validate(now time.Time) error {
	// Default dates
	if r.From.IsZero() {
		r.From = now.AddDate(0, 0, -30)
	}
	if r.To.IsZero() {
		r.To = now
	}

	log.Println("from date :", r.From)
	log.Println("to date :", r.To)
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
	return nil
}

// if all values are zero, then default default to daily interval of last 30 days
func (r *RevenueAnalyticsRequest) Validate(now time.Time) error {
	// Default dates
	if r.From.IsZero() {
		r.From = now.AddDate(0, 0, -30)
	}
	if r.To.IsZero() {
		r.To = now
	}

	if r.Interval == "" {
		r.Interval = "daily"
	}
	if r.Interval != "daily" && r.Interval != "weekly" && r.Interval != "monthly" && r.Interval != "yearly" {
		return errors.New("invalid interval")
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
	return nil
}

type DashboardRequest struct {
	From     time.Time `form:"from" time_format:"2006-01-02"`
	To       time.Time `form:"to" time_format:"2006-01-02"`
	Interval string    `form:"interval"` // day | week | month
}

type DashboardResponse struct {
	From time.Time `json:"from,omitempty"`
	To   time.Time `json:"to,omitempty"`

	Stats *DashboardStats `json:"stats"`

	Summary       *SalesSummary `json:"summary"`
	TopProducts   []TopStatItem `json:"top_products"`
	TopCategories []TopStatItem `json:"top_categories"`
	TopBrands     []TopStatItem `json:"top_brands"`
}

type DashboardStats struct {
	UserStats    *UserStats    `json:"user_stats"`
	OrderStats   *OrderStats   `json:"order_stats"`
	ProductStats *ProductStats `json:"product_stats"`

	ActiveCoupons int64 `json:"active_coupons"`
}

type UserStats struct {
	TotalUsers       int64 `json:"total_users"`
	TotalActiveUsers int64 `json:"total_active_users"`
}
type OrderStats struct {
	TotalOrders          int64 `json:"total_orders"`
	TotalPendingOrders   int64 `json:"total_pending_orders"`
	TotalDeliveredOrders int64 `json:"total_delivered_orders"`
	TotalCancelledOrders int64 `json:"total_cancelled_orders"`
	TotalReturnedOrders  int64 `json:"total_returned_orders"`
}

type ProductStats struct {
	TotalProducts   int64 `json:"total_products"`
	TotalCategories int64 `json:"total_categories"`
	TotalBrands     int64 `json:"total_brands"`
}

func (r *DashboardRequest) Validate(now time.Time) error {
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

	if r.Interval == "" {
		r.Interval = "day"
	}
	if r.Interval != "day" && r.Interval != "week" && r.Interval != "month" {
		return errors.New("invalid interval")
	}
	return nil
}
