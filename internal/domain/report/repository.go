package report

import (
	"context"
)

type ReportRepository interface {
	GetSalesReport(ctx context.Context, req *SalesReportRequest) (SalesReportResponse, error)

	GetTopSellingProducts(ctx context.Context, req *TopSellingRequest) ([]TopStatItem, error)
	GetTopSellingCategories(ctx context.Context, req *TopSellingRequest) ([]TopStatItem, error)
	GetTopSellingBrands(ctx context.Context, req *TopSellingRequest) ([]TopStatItem, error)

	GetRevenueAnalytics(ctx context.Context, req *RevenueAnalyticsRequest) (*RevenueAnalyticsResponse, error)
}
