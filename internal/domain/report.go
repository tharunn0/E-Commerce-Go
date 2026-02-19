package domain

import (
	"context"
	"time"
)

type ReportRepository interface {
	GetSalesReport(ctx context.Context, req *SalesReportRequest) (SalesReportResponse, error)
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
	CouponDiscount float64 `json:"coupon_discount"`
	NetRevenue     float64 `json:"net_revenue"`
}
