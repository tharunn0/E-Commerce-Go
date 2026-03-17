package repository

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/report"
)

type ReportRepository struct {
	db *pgxpool.Pool
}

func NewReportRepository(db *pgxpool.Pool) *ReportRepository {
	return &ReportRepository{
		db: db,
	}
}

func (r *ReportRepository) GetSalesReport(ctx context.Context, req *report.SalesReportRequest) (report.SalesReportResponse, error) {

	query := `SELECT
    COUNT(DISTINCT o.id) AS total_orders,

    COALESCE(SUM(
        CASE
            WHEN oi.status = 'delivered'
            THEN oi.total_price
            ELSE 0
        END
    ), 0) AS delivered_revenue,

    COALESCE(SUM(
        CASE
            WHEN oi.status = 'returned'
            THEN oi.total_price
            ELSE 0
        END
    ), 0) AS returned_amount,

    COALESCE(SUM(oc.discount_applied), 0) AS coupon_discount

FROM orders o
JOIN order_items oi ON oi.order_id = o.id
LEFT JOIN order_coupons oc ON oc.order_id = o.id
WHERE
    o.status IN ('delivered', 'shipped')
    AND o.created_at >= $1
    AND o.created_at <= $2;

`

	row := r.db.QueryRow(ctx, query, req.From, req.To)

	var sales report.SalesReportResponse
	err := row.Scan(&sales.TotalOrders, &sales.GrossRevenue, &sales.ReturnedAmount, &sales.CouponDiscount)
	if err != nil {
		return report.SalesReportResponse{}, err
	}

	log.Println("Gross Revenue: ", sales.GrossRevenue)
	log.Println("Returned Amount: ", sales.ReturnedAmount)
	log.Println("Coupon Discount: ", sales.CouponDiscount)

	// calculate net revenue
	sales.NetRevenue = sales.GrossRevenue - sales.CouponDiscount - sales.ReturnedAmount

	return sales, nil
}

func (r *ReportRepository) GetTopSellingProducts(ctx context.Context, req *report.TopSellingRequest) ([]report.TopStatItem, error) {

	query := `SELECT
    p.id,
    p.name,
    SUM(oi.quantity) AS total_sold,
    SUM(oi.total_price) AS total_revenue
FROM order_items oi
JOIN product_variants pv 
    ON pv.id = oi.product_variant_id
JOIN products p
    ON p.id = pv.product_id
JOIN orders o
    ON o.id = oi.order_id
WHERE
    oi.status = 'delivered'
    AND o.status IN ('delivered', 'shipped')
    AND o.created_at >= $1
    AND o.created_at <= $2
GROUP BY
    p.id,
    p.name
ORDER BY
    total_sold DESC
LIMIT $3;
`

	log.Println("Limit: ", req.Limit)

	rows, err := r.db.Query(ctx, query, req.From, req.To, req.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []report.TopStatItem
	for rows.Next() {
		var item report.TopStatItem
		err := rows.Scan(&item.ID, &item.Name, &item.TotalSold, &item.TotalRevenue)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func (r *ReportRepository) GetTopSellingCategories(ctx context.Context, req *report.TopSellingRequest) ([]report.TopStatItem, error) {

	query := `SELECT
    	c.id,
    	c.name,
    	SUM(oi.quantity) AS total_sold,
    	SUM(oi.total_price) AS total_revenue
	FROM order_items oi
	JOIN product_variants pv 
    	ON pv.id = oi.product_variant_id
	JOIN products p
    	ON p.id = pv.product_id
	JOIN categories c
    	ON c.id = p.category_id
	JOIN orders o
    	ON o.id = oi.order_id
	WHERE
    	oi.status = 'delivered'
    	AND o.status IN ('delivered', 'shipped')
    	AND o.created_at >= $1
    	AND o.created_at <= $2
	GROUP BY
    	c.id,
    	c.name
	ORDER BY
    	total_sold DESC,
    	total_revenue DESC
	LIMIT $3;
`

	log.Println("Limit: ", req.Limit)

	rows, err := r.db.Query(ctx, query, req.From, req.To, req.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []report.TopStatItem
	for rows.Next() {
		var item report.TopStatItem
		err := rows.Scan(&item.ID, &item.Name, &item.TotalSold, &item.TotalRevenue)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func (r *ReportRepository) GetTopSellingBrands(ctx context.Context, req *report.TopSellingRequest) ([]report.TopStatItem, error) {

	query := `SELECT
    	b.id,
    	b.name,
    	SUM(oi.quantity) AS total_sold,
    	SUM(oi.total_price) AS total_revenue
	FROM order_items oi
	JOIN product_variants pv 
    	ON pv.id = oi.product_variant_id
	JOIN products p
    	ON p.id = pv.product_id
	JOIN brands b
    	ON b.id = p.brand_id
	JOIN orders o
    	ON o.id = oi.order_id
	WHERE
    	oi.status = 'delivered'
    	AND o.status IN ('delivered', 'shipped')
    	AND o.created_at >= $1
    	AND o.created_at <= $2
	GROUP BY
    	b.id,
    	b.name
	ORDER BY
    	total_sold DESC,
    	total_revenue DESC
	LIMIT $3;
`

	log.Println("Limit: ", req.Limit)

	rows, err := r.db.Query(ctx, query, req.From, req.To, req.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []report.TopStatItem
	for rows.Next() {
		var item report.TopStatItem
		err := rows.Scan(&item.ID, &item.Name, &item.TotalSold, &item.TotalRevenue)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func (r *ReportRepository) GetRevenueAnalytics(ctx context.Context, req *report.RevenueAnalyticsRequest) (*report.RevenueAnalyticsResponse, error) {

	var intervalUnit string
	var intervalSQL string
	var seriesInterval string

	switch req.Interval {
	case "weekly":
		intervalUnit = "week"
		intervalSQL = "date_trunc('week', created_at)"
		seriesInterval = "1 week"

	case "monthly":
		intervalUnit = "month"
		intervalSQL = "date_trunc('month', created_at)"
		seriesInterval = "1 month"

	default:
		intervalUnit = "day"
		intervalSQL = "date_trunc('day', created_at)"
		seriesInterval = "1 day"
	}

	query := fmt.Sprintf(`
WITH series AS (
	SELECT generate_series(
		date_trunc('%s', $1::timestamptz),
		date_trunc('%s', $2::timestamptz),
		'%s'
	) AS period
),
revenue AS (
	SELECT
		%s AS period,
		SUM(total_amount) AS revenue
	FROM orders
	WHERE created_at BETWEEN $1 AND $2
	AND status = 'delivered'
	GROUP BY period
)
SELECT
	s.period,
	COALESCE(r.revenue, 0) AS revenue
FROM series s
LEFT JOIN revenue r ON r.period = s.period
ORDER BY s.period;
`,
		intervalUnit,
		intervalUnit,
		seriesInterval,
		intervalSQL,
	)

	rows, err := r.db.Query(ctx, query, req.From, req.To)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var data []report.RevenueData

	for rows.Next() {
		var d report.RevenueData
		if err := rows.Scan(&d.Date, &d.Revenue); err != nil {
			return nil, err
		}
		data = append(data, d)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &report.RevenueAnalyticsResponse{
		RevenueAnalyticsRequest: *req,
		RevenueData:             data,
	}, nil
}
