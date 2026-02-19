package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tharunn0/E-Commerce-Go/internal/domain"
)

type ReportRepository struct {
	db *pgxpool.Pool
}

func NewReportRepository(db *pgxpool.Pool) *ReportRepository {
	return &ReportRepository{
		db: db,
	}
}

func (r *ReportRepository) GetSalesReport(ctx context.Context, req *domain.SalesReportRequest) (domain.SalesReportResponse, error) {

	query := `SELECT
    COUNT(o.id) AS total_orders,
    COALESCE(SUM(o.total_amount), 0) AS gross_revenue,
    COALESCE(SUM(oc.discount_applied), 0) AS coupon_discount
FROM orders o
LEFT JOIN order_coupons oc ON oc.order_id = o.id
WHERE
    o.status = 'delivered'
    AND o.created_at >= $1
    AND o.created_at <= $2;
`

	row := r.db.QueryRow(ctx, query, req.FromDate, req.ToDate)

	var sales domain.SalesReportResponse
	err := row.Scan(&sales.TotalOrders, &sales.GrossRevenue, &sales.CouponDiscount)
	if err != nil {
		return domain.SalesReportResponse{}, err
	}
	return sales, nil
}
