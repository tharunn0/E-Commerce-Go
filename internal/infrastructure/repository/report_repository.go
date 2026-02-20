package repository

import (
	"context"
	"log"

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

	row := r.db.QueryRow(ctx, query, req.FromDate, req.ToDate)

	var sales domain.SalesReportResponse
	err := row.Scan(&sales.TotalOrders, &sales.GrossRevenue, &sales.ReturnedAmount, &sales.CouponDiscount)
	if err != nil {
		return domain.SalesReportResponse{}, err
	}

	log.Println("Gross Revenue: ", sales.GrossRevenue)
	log.Println("Returned Amount: ", sales.ReturnedAmount)
	log.Println("Coupon Discount: ", sales.CouponDiscount)

	// calculate net revenue
	sales.NetRevenue = sales.GrossRevenue - sales.CouponDiscount - sales.ReturnedAmount

	return sales, nil
}

func (r *ReportRepository) GetTopSellingProducts(ctx context.Context, req *domain.TopSellingRequest) ([]domain.TopStatItem, error) {

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

	var items []domain.TopStatItem
	for rows.Next() {
		var item domain.TopStatItem
		err := rows.Scan(&item.ID, &item.Name, &item.TotalSold, &item.TotalRevenue)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func (r *ReportRepository) GetTopSellingCategories(ctx context.Context, req *domain.TopSellingRequest) ([]domain.TopStatItem, error) {

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

	var items []domain.TopStatItem
	for rows.Next() {
		var item domain.TopStatItem
		err := rows.Scan(&item.ID, &item.Name, &item.TotalSold, &item.TotalRevenue)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func (r *ReportRepository) GetTopSellingBrands(ctx context.Context, req *domain.TopSellingRequest) ([]domain.TopStatItem, error) {

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

	var items []domain.TopStatItem
	for rows.Next() {
		var item domain.TopStatItem
		err := rows.Scan(&item.ID, &item.Name, &item.TotalSold, &item.TotalRevenue)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}
