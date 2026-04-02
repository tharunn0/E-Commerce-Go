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

	query := `
SELECT
    COUNT(DISTINCT o.id) AS total_orders,

    COALESCE(SUM(
        CASE WHEN oi.status = 'delivered'
        THEN oi.total_price ELSE 0 END
    ),0) AS gross_revenue,

    COALESCE(SUM(
        CASE WHEN oi.status = 'returned'
        THEN oi.total_price ELSE 0 END
    ),0) AS returned_amount,

    COALESCE((
        SELECT SUM(discount_applied)
        FROM order_coupons oc
        JOIN orders o2 ON o2.id = oc.order_id
        WHERE o2.status = 'delivered'
        AND o2.created_at BETWEEN $1 AND $2
    ),0) AS coupon_discount

FROM orders o
JOIN order_items oi ON oi.order_id = o.id
WHERE
    o.status = 'delivered'
    AND o.created_at BETWEEN $1 AND $2;
`

	var sales report.SalesReportResponse

	err := r.db.QueryRow(ctx, query, req.From, req.To).Scan(
		&sales.TotalOrders,
		&sales.GrossRevenue,
		&sales.ReturnedAmount,
		&sales.CouponDiscount,
	)
	if err != nil {
		return report.SalesReportResponse{}, err
	}

	sales.NetRevenue =
		sales.GrossRevenue -
			sales.ReturnedAmount -
			sales.CouponDiscount

	sales.From = req.From
	sales.To = req.To

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

func (r *ReportRepository) GetDashboard(ctx context.Context, req *report.DashboardRequest) (*report.DashboardResponse, error) {

	resp := report.DashboardResponse{}

	// From time.Time `json:"from,omitempty"`
	// To   time.Time `json:"to,omitempty"`

	// Stats DashboardStats `json:"stats"`

	// Summary       SalesSummary  `json:"summary"`
	// RevenueChart  []RevenueData `json:"revenue_chart"`
	// TopProducts   []TopStatItem `json:"top_products"`
	// TopCategories []TopStatItem `json:"top_categories"`
	// TopBrands     []TopStatItem `json:"top_brands"`

	stats := report.DashboardStats{}

	// get user stats
	userStat := report.UserStats{}
	query := `SELECT
					COUNT(*) as total_users,
					COUNT(*) FILTER (WHERE is_verified = true) as total_active_users
				FROM users
				WHERE role = 'user';`
	err := r.db.QueryRow(ctx, query).Scan(&userStat.TotalUsers, &userStat.TotalActiveUsers)
	if err != nil {
		return nil, err
	}
	stats.UserStats = &userStat

	// get order stats
	orderStat := report.OrderStats{}
	query = `SELECT
					COUNT(*) as total_orders,
					COUNT(*) FILTER (WHERE status = 'pending') as total_pending_orders,
					COUNT(*) FILTER (WHERE status = 'delivered') as total_delivered_orders,
					COUNT(*) FILTER (WHERE status = 'cancelled') as total_cancelled_orders,
					COUNT(*) FILTER (WHERE return_status = 'returned') as total_returned_orders
				FROM orders;`
	err = r.db.QueryRow(ctx, query).Scan(&orderStat.TotalOrders, &orderStat.TotalPendingOrders, &orderStat.TotalDeliveredOrders, &orderStat.TotalCancelledOrders, &orderStat.TotalReturnedOrders)
	if err != nil {
		return nil, err
	}
	stats.OrderStats = &orderStat

	// get product stats
	productStat := report.ProductStats{}
	query = `SELECT
					(SELECT COUNT(*) as total_products FROM products),
					(SELECT COUNT(*) as total_categories FROM categories),
					(SELECT COUNT(*) as total_brands FROM brands);`
	err = r.db.QueryRow(ctx, query).Scan(&productStat.TotalProducts, &productStat.TotalCategories, &productStat.TotalBrands)
	if err != nil {
		return nil, err
	}
	stats.ProductStats = &productStat

	resp.Stats = &stats

	// get active coupon count
	var activeCouponCount int64
	query = `SELECT COUNT(*) as active_coupon_count FROM coupons
			WHERE is_active = true;`
	err = r.db.QueryRow(ctx, query).Scan(&activeCouponCount)
	if err != nil {
		return nil, err
	}
	stats.ActiveCoupons = activeCouponCount

	resp.Stats = &stats

	return &resp, nil
}
