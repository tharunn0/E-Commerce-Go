package repository

import (
	// "context"

	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tharunn0/E-Commerce-Go/internal/apperror"
	"github.com/tharunn0/E-Commerce-Go/internal/domain"
)

type OrderRepository struct {
	DB *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) OrderRepository {
	return OrderRepository{DB: db}
}

func (r OrderRepository) CreateOrder(ctx context.Context, data *domain.CreateOrderData) error {

	var query string

	// begin transaction
	tx, err := r.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		fmt.Println("error beginning transaction", err)
		return err
	}

	// rollback on error
	defer tx.Rollback(ctx)

	// check if enough stock is available
	for _, item := range data.Items {
		query = `SELECT stock FROM product_variants WHERE id = $1`
		var stock int64
		err = r.DB.QueryRow(ctx, query, item.ProductVariantID).Scan(&stock)
		if err != nil {
			return err
		}
		if stock < item.Quantity {
			return apperror.ErrNotEnoughStock
		}
	}

	var internalOrderID int64
	// insert into orders
	query = `INSERT INTO orders (user_id, public_order_id, total_amount, tax_amount, shipping_address_id, billing_address_id,estimated_delivery_date,status)
	VALUES ($1, $2, $3, $4, $5, $6,$7,$8) RETURNING id`
	err = tx.QueryRow(ctx, query, data.UserID, data.OrderID, data.TotalAmount, data.TaxAmount,
		data.ShippingAddressID, data.BillingAddressID, data.EstimatedDeliveryDate, strings.ToLower(string(data.Status))).Scan(&internalOrderID)
	if err != nil {
		fmt.Println("error inserting order", err)
		return err
	}

	// insert into order_items
	for _, item := range data.Items {
		query = `INSERT INTO order_items (order_id, product_variant_id, sku_at_purchase, product_name_at_purchase, quantity, unit_price, total_price,status)
	SELECT $1, $2, pv.sku, p.name, $3, $4, $5, $6 FROM product_variants pv
	JOIN products p ON pv.product_id = p.id WHERE pv.id = $2`
		_, err = tx.Exec(ctx, query, internalOrderID, item.ProductVariantID, item.Quantity, item.UnitPrice, item.TotalPrice, "pending")
		if err != nil {
			fmt.Println("error inserting order items", err)
			return err
		}
	}

	// inset into shipments
	query = `INSERT INTO shipments (order_id, status,type) VALUES
	 ($1, $2,$3)`
	_, err = tx.Exec(ctx, query, internalOrderID, "pending", data.DeliveryType)
	if err != nil {
		fmt.Println("error inserting shipment", err)
		return err
	}

	// update product variant stock
	for _, item := range data.Items {
		query = `UPDATE product_variants SET stock = stock - $1 WHERE id = $2`
		_, err = tx.Exec(ctx, query, item.Quantity, item.ProductVariantID)
		if err != nil {
			fmt.Println("error updating product variant stock", err)
			return err
		}
	}

	// empty cart
	if data.OrderSource == "cart" {
		query = `DELETE FROM cart_items 
	 	USING carts C 
	 	WHERE cart_items.cart_id = C.id AND C.user_id = $1`
		_, err = tx.Exec(ctx, query, data.UserID)
		if err != nil {
			fmt.Println("error emptying cart", err)
			return err
		}
	}

	// check if total amount of items is greater that 10_00_000
	if data.TotalAmount > 10_00_000 {
		return apperror.ErrOrderAmountExceeded
	}

	// commit transaction
	return tx.Commit(ctx)
}

func (r OrderRepository) GetUserOrders(ctx context.Context, userID int64) ([]domain.OrderBaseResponse, error) {

	query := `SELECT o.public_order_id, o.total_amount, o.tax_amount, o.status,o.return_status,o.estimated_delivery_date,p.currency,o.shipping_address_id,
		s.type as delivery_type,s.status as shipment_status,p.status as payment_status,
		a.id, a.label,a.address_line,a.address_line_2,a.city,a.district,a.state,a.pincode,a.country
		FROM orders o 
		LEFT JOIN shipments s ON o.id = s.order_id
		LEFT JOIN payments p ON o.id = p.order_id
		LEFT JOIN user_addresses a ON o.shipping_address_id = a.id
		WHERE o.user_id = $1`
	rows, err := r.DB.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []domain.OrderBaseResponse
	for rows.Next() {
		var order domain.OrderBaseResponse
		var address domain.UserAddress
		err = rows.Scan(&order.OrderID, &order.TotalAmount, &order.TaxAmount, &order.Status, &order.ReturnStatus, &order.EstimatedDeliveryDate, &order.Currency, &order.ShippingAddressID,
			&order.DeliveryType, &order.ShipmentStatus, &order.PaymentStatus, &address.ID, &address.Label, &address.AddressLine, &address.AddressLine2, &address.City,
			&address.District, &address.State, &address.Pincode, &address.Country)
		if err != nil {
			return nil, err
		}
		order.Subtotal = order.TotalAmount - order.TaxAmount
		order.ShippingAddress = &address
		orders = append(orders, order)
	}

	return orders, nil
}

func (r OrderRepository) GetUserOrderByID(ctx context.Context, orderID string, userID int64) (*domain.OrderResponse, error) {
	query := `SELECT o.public_order_id, o.total_amount, o.tax_amount, o.status,o.return_status,o.estimated_delivery_date,p.currency,o.shipping_address_id,o.billing_address_id,
		s.type as delivery_type,s.status as shipment_status,p.status as payment_status,p.provider,o.created_at,o.updated_at
		FROM orders o 
		LEFT JOIN shipments s ON o.id = s.order_id
		LEFT JOIN payments p ON o.id = p.order_id
		WHERE o.public_order_id = $1`

	if userID != 0 {
		query += fmt.Sprintf(" AND o.user_id = %d", userID)
	}

	var order domain.OrderResponse
	row := r.DB.QueryRow(ctx, query, orderID)
	err := row.Scan(&order.OrderID, &order.Subtotal, &order.TaxAmount, &order.Status, &order.ReturnStatus, &order.EstimatedDeliveryDate, &order.Currency,
		&order.ShippingAddressID, &order.BillingAddressID,
		&order.DeliveryType, &order.ShipmentStatus, &order.PaymentStatus, &order.PaymentMethod,
		&order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		fmt.Println("query failed", query, orderID)
		if err == pgx.ErrNoRows {
			return nil, apperror.ErrOrderNotFound
		}
		return nil, err
	}

	items := []domain.OrderItem{}
	query = `SELECT oi.id, pv.id, p.name, pv.sku, oi.quantity, oi.unit_price, oi.status,oi.total_price,pvi.url
		FROM order_items oi
		LEFT JOIN product_variants pv ON oi.product_variant_id = pv.id
		LEFT JOIN products p ON pv.product_id = p.id
		LEFT JOIN product_variant_images pvi ON pv.id = pvi.product_variant_id
		WHERE oi.order_id = (SELECT id FROM orders WHERE public_order_id = $1)
		`
	rows, err := r.DB.Query(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item domain.OrderItem
		err = rows.Scan(&item.ItemID, &item.ProductVariantID, &item.ProductName, &item.SKU, &item.Quantity, &item.UnitPrice, &item.Status, &item.TotalPrice, &item.ImageURL)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	order.Items = items

	// change payment_status to refunded if order status is returned
	if order.ReturnStatus != nil && *order.ReturnStatus == "returned" {
		order.PaymentStatus = "refunded"
	}

	// order.Subtotal = order.TotalAmount - order.TaxAmount
	return &order, nil
}

func (r OrderRepository) UpdateOrderStatusOnPayment(ctx context.Context, orderID string, reqStatus domain.PaymentStatus) error {

	var status string
	if reqStatus == domain.PaymentStatusCompleted {
		status = "confirmed"
	} else {
		status = "failed"
	}

	query := `UPDATE orders o 
	SET status = $1, updated_at = now()
	 FROM payments p 
	WHERE o.id = p.order_id AND p.provider_order_id = $2`
	cmdTag, err := r.DB.Exec(ctx, query, status, orderID)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return apperror.ErrPaymentNotFound
	}

	query = `UPDATE order_items oi
	SET status = $1 
	FROM payments p
	WHERE oi.order_id = p.order_id AND p.provider_order_id = $2`
	_, err = r.DB.Exec(ctx, query, status, orderID)
	if err != nil {
		return err
	}

	// if payment failed, update stocks
	if status == "failed" {
		query = `WITH failed_order AS (
			SELECT o.id FROM orders o 
			JOIN payments p ON o.id = p.order_id
			WHERE p.provider_order_id = $1
			AND o.status IN ('pending', 'failed', 'cancelled')
		)
		UPDATE product_variants pv
		SET stock = pv.stock + oi.quantity
		FROM order_items oi, failed_order fo
		WHERE oi.product_variant_id = pv.id
		  AND oi.order_id = fo.id;`
		_, err = r.DB.Exec(ctx, query, orderID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r OrderRepository) UpdateShipmentStatus(ctx context.Context, orderID string, status domain.ShipmentStatus, cod bool) error {

	statusString := string(status) // will be shipped or delivered

	tx, err := r.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	var query string
	if statusString == "shipped" {
		query = `UPDATE shipments
	 SET status = $1, shipped_at = now()
	 FROM orders o
	 WHERE o.id = shipments.order_id AND o.public_order_id = $2`
	} else {
		query = `UPDATE shipments
	 SET status = $1, delivered_at = now()
	 FROM orders o
	 WHERE o.id = shipments.order_id AND o.public_order_id = $2`
	}

	cmdTag, err := r.DB.Exec(ctx, query, statusString, orderID)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return apperror.ErrOrderNotFound
	}

	query = `UPDATE orders
	 SET status = $1
	 WHERE public_order_id = $2`
	cmdTag, err = r.DB.Exec(ctx, query, statusString, orderID)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return apperror.ErrOrderNotFound
	}

	var pstatus string
	if statusString == "delivered" {
		pstatus = "delivered"
	} else {
		pstatus = "shipped"
	}

	query = `UPDATE order_items oi
	SET status = $1
	FROM orders o
	WHERE oi.order_id = o.id AND o.public_order_id = $2`
	cmdTag, err = r.DB.Exec(ctx, query, pstatus, orderID)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return apperror.ErrOrderItemNotFound
	}

	if cod && statusString == "delivered" {
		query = `UPDATE payments
		 SET status = 'paid', paid_at = now()
		 FROM orders o
		 WHERE o.id = payments.order_id AND o.public_order_id = $1`
		cmdTag, err = r.DB.Exec(ctx, query, orderID)
		if err != nil {
			return err
		}
		if cmdTag.RowsAffected() == 0 {
			return apperror.ErrOrderItemNotFound
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func (r OrderRepository) CancelOrder(ctx context.Context, orderID string, reason string, orderItemIDs []int64) error {

	// Start transaction
	tx, err := r.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	// Ensure rollback on any early return or panic
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// Lock the order row to prevent concurrent cancellations
	var internalOrderID int64
	query := `
		SELECT id
		FROM orders
		WHERE public_order_id = $1
		FOR UPDATE
	`
	if err = tx.QueryRow(ctx, query, orderID).Scan(&internalOrderID); err != nil {
		return err
	}

	// FULL ORDER CANCELLATION
	if len(orderItemIDs) == 0 {

		// Cancel all order items (idempotent)
		_, err = tx.Exec(ctx, `
			UPDATE order_items
			SET status = 'cancelled'
			WHERE order_id = $1
			  AND status != 'cancelled'
		`, internalOrderID)
		if err != nil {
			return err
		}

		// Cancel order
		_, err = tx.Exec(ctx, `
			UPDATE orders
			SET status = 'cancelled',
			    updated_at = now()
			WHERE id = $1
			  AND status != 'cancelled'
		`, internalOrderID)
		if err != nil {
			return err
		}

		// Cancel shipment (only if not already terminal)
		_, err = tx.Exec(ctx, `
			UPDATE shipments
			SET status = 'cancelled',
			    updated_at = now()
			WHERE order_id = $1
			  AND status NOT IN ('shipped', 'delivered', 'cancelled')
		`, internalOrderID)
		if err != nil {
			return err
		}

		// Update payment state (assumes not captured yet)
		_, err = tx.Exec(ctx, `
			UPDATE payments
			SET status = 'cancelled',
			    updated_at = now()
			WHERE order_id = $1
			  AND status = 'pending'
		`, internalOrderID)
		if err != nil {
			return err
		}

		// Commit transaction
		return tx.Commit(ctx)
	}

	// Cancel only selected order items (idempotent)
	_, err = tx.Exec(ctx, `
		UPDATE order_items
		SET status = 'cancelled'
		WHERE order_id = $1
		  AND id = ANY($2)
		  AND status != 'cancelled'
	`, internalOrderID, orderItemIDs)
	if err != nil {
		return err
	}

	// Check if ALL order items are now cancelled
	var remainingActiveItems int
	err = tx.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM order_items
		WHERE order_id = $1
		  AND status != 'cancelled'
	`, internalOrderID).Scan(&remainingActiveItems)
	if err != nil {
		return err
	}

	// If all items are cancelled, cascade cancellation
	if remainingActiveItems == 0 {

		// Cancel order
		_, err = tx.Exec(ctx, `
			UPDATE orders
			SET status = 'cancelled',
			    updated_at = now()
			WHERE id = $1
			  AND status != 'cancelled'
		`, internalOrderID)
		if err != nil {
			return err
		}

		// Cancel shipment (only if not shipped)
		_, err = tx.Exec(ctx, `
			UPDATE shipments
			SET status = 'cancelled',
			    updated_at = now()
			WHERE order_id = $1
			  AND status NOT IN ('shipped', 'delivered', 'cancelled')
		`, internalOrderID)
		if err != nil {
			return err
		}

		// Update payment state safely
		_, err = tx.Exec(ctx, `
			UPDATE payments
			SET status = 'cancelled',
			    updated_at = now()
			WHERE order_id = $1
			  AND status = 'pending'
		`, internalOrderID)
		if err != nil {
			return err
		}
	}

	// commit
	return tx.Commit(ctx)
}

// returns
func (r OrderRepository) ReturnOrderRequest(ctx context.Context, orderID string, userID int64, reason string) error {
	// order_return table
	// id | order_id | user_id | refunded_amount | status | created_at
	// return_items table
	// id | return_id | order_item_id | quantity | status | reason | requested_at | approved_at | order_item_price

	tx, err := r.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	query := `SELECT id FROM orders WHERE public_order_id = $1`
	var internalOrderID int64
	err = tx.QueryRow(ctx, query, orderID).Scan(&internalOrderID)
	if err != nil {
		fmt.Println("failed to fetch order return request", err, query)
		return err
	}

	// insert the order return request , once for every order
	query = `INSERT INTO order_returns (order_id,user_id,refunded_amount,status,created_at)
	VALUES ($1,$2,$3,$4,$5)
	ON CONFLICT (order_id)
	DO UPDATE SET order_id = EXCLUDED.order_id
	RETURNING id`
	var returnID int64
	err = tx.QueryRow(ctx, query, internalOrderID, userID, 0, "requested", time.Now()).Scan(&returnID)
	if err != nil {
		fmt.Println("failed to insert order return request", err, query)
		return err
	}

	// insert the return items
	query = `INSERT INTO return_items (return_id,order_item_id,quantity,order_item_price,status,reason,requested_at)
	SELECT 
	 $1 as return_id,
	 id as order_item_id, 
	 quantity, 
	 total_price as order_item_price, 
	 'requested' as status, 
	 $2 as reason, 
	 now() as requested_at 
	 FROM order_items WHERE order_id = $3`
	_, err = tx.Exec(ctx, query, returnID, reason, internalOrderID)
	if err != nil {
		fmt.Println("failed to insert return items", err, query)
		return err
	}

	// update order status
	query = `UPDATE orders SET return_status = 'requested' WHERE id = $1`
	cmdTag, err := tx.Exec(ctx, query, internalOrderID)
	if err != nil {
		fmt.Println("failed to update order status", err, query)
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		fmt.Println("failed to update order status", cmdTag.RowsAffected(), query)
	}
	fmt.Println("return request created successfully", internalOrderID)

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func (r OrderRepository) ReturnOrderItemRequest(ctx context.Context, orderID string, userID int64, itemID int64, reason string) error {
	// Use a transaction to ensure consistency
	tx, err := r.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	fmt.Println("transaction started")
	defer tx.Rollback(ctx) // Will be ignored if commit succeeds

	var internalOrderID int64
	query := `SELECT id FROM orders WHERE public_order_id = $1`
	err = tx.QueryRow(ctx, query, orderID).Scan(&internalOrderID)
	if err != nil {
		fmt.Println("failed to fetch internal order id", err)
		return fmt.Errorf("order not found: %w", err)
	}

	var exist bool
	query = `
		SELECT EXISTS(
			SELECT 1 
			FROM order_items oi
			JOIN orders o ON oi.order_id = o.id
			WHERE oi.id = $1 
			  AND oi.order_id = $2 
			  AND o.user_id = $3
		)`
	err = tx.QueryRow(ctx, query, itemID, internalOrderID, userID).Scan(&exist)
	if err != nil {
		fmt.Println("failed to verify order item ownership", err)
		return fmt.Errorf("failed to verify order item ownership: %w", err)
	}
	if !exist {
		return apperror.ErrReturnItemNotFound
	}

	var returnID int64
	query = `
		INSERT INTO order_returns (order_id, user_id, refunded_amount, status, created_at)
		VALUES ($1, $2, 0, 'requested', NOW())
		ON CONFLICT (order_id) DO UPDATE 
		SET status = 'requested'
		RETURNING id`
	err = tx.QueryRow(ctx, query, internalOrderID, userID).Scan(&returnID)
	if err != nil {
		fmt.Println("failed to upsert order_returns", err)
		return fmt.Errorf("failed to create/update return request: %w", err)
	}

	query = `
		INSERT INTO return_items 
			(return_id, order_item_id, quantity, order_item_price, status, reason, requested_at)
		SELECT 
			$1, 
			oi.id, 
			oi.quantity, 
			oi.total_price, 
			'requested', 
			$2, 
			NOW()
		FROM order_items oi
		WHERE oi.id = $3`
	cmdTag, err := tx.Exec(ctx, query, returnID, reason, itemID)
	if err != nil {
		return fmt.Errorf("failed to insert return item: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("insertion failed for return item")
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	fmt.Println("transaction committed successfully")

	return nil
}

func (r OrderRepository) ListAllReturns(ctx context.Context, filter *domain.ReturnFilter) ([]domain.BaseReturnResponse, error) {

	base := `
	SELECT
		ors.id,
		o.public_order_id,
		ors.status,
		ors.refunded_amount,
		ors.created_at,
		COALESCE(COUNT(ri.id), 0) AS total_items,
        COALESCE(SUM(ri.order_item_price), 0) AS total_refund_value
	FROM order_returns ors
	JOIN orders o ON ors.order_id = o.id
	LEFT JOIN return_items ri ON ri.return_id = ors.id
	`

	conditions := []string{}
	args := []any{}
	argPos := 1

	if filter.OrderID != nil {
		conditions = append(conditions, fmt.Sprintf("o.public_order_id = $%d", argPos))
		args = append(args, *filter.OrderID)
		argPos++
	}

	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("ors.status = $%d", argPos))
		args = append(args, *filter.Status)
		argPos++
	}

	if filter.CreatedFrom != nil {
		conditions = append(conditions, fmt.Sprintf("ors.created_at >= $%d", argPos))
		args = append(args, *filter.CreatedFrom)
		argPos++
	}

	if filter.CreatedTo != nil {
		conditions = append(conditions, fmt.Sprintf("ors.created_at <= $%d", argPos))
		args = append(args, *filter.CreatedTo)
		argPos++
	}

	query := base

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " GROUP BY ors.id, o.public_order_id, ors.refunded_amount, ors.created_at"

	if filter.OrderBy != nil {
		query += " ORDER BY " + *filter.OrderBy
		if filter.Sort != nil {
			query += " " + *filter.Sort
		}
	}

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filter.Limit)
	}

	if filter.Page > 0 && filter.Limit > 0 {
		offset := (filter.Page - 1) * filter.Limit
		query += fmt.Sprintf(" OFFSET %d", offset)
	}

	rows, err := r.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []domain.BaseReturnResponse{}
	for rows.Next() {
		var r domain.BaseReturnResponse
		err := rows.Scan(
			&r.ID,
			&r.OrderID,
			&r.Status,
			&r.RefundedAmount,
			&r.CreatedAt,
			&r.TotalItems,
			&r.TotalRefundValue,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, r)
	}

	return results, nil
}

func (r OrderRepository) GetReturnRequest(ctx context.Context, returnID int64) (*domain.FullReturnResponse, error) {

	mainQuery := `
        SELECT 
            r.id,
            r.order_id,
            r.user_id AS "user.id",
            u.first_name AS "user.name",
            u.email AS "user.email",
            r.status,
            r.refunded_amount,
            r.created_at,
            COUNT(ri.id) AS total_items,
            COALESCE(SUM(ri.quantity), 0) AS total_quantity,
            COALESCE(SUM(ri.order_item_price), 0) AS total_refundable_amount
        FROM order_returns r
        LEFT JOIN users u ON r.user_id = u.id 
        LEFT JOIN return_items ri ON ri.return_id = r.id
        WHERE r.id = $1
        GROUP BY r.id, r.order_id, r.user_id, u.first_name, u.email, r.status, r.refunded_amount, r.created_at`

	var result domain.FullReturnResponse

	err := r.DB.QueryRow(ctx, mainQuery, returnID).Scan(
		&result.ID,
		&result.OrderID,
		&result.User.ID,    // "user.id"
		&result.User.Name,  // "user.name" -> first_name
		&result.User.Email, // "user.email"
		&result.Status,
		&result.RefundedAmount,
		&result.CreatedAt,
		&result.TotalItems,
		&result.TotalQuantity,
		&result.TotalRefundValue,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperror.ErrReturnRequestNotFound
		}
		return nil, err
	}

	// Items query: fetch all return items
	itemsQuery := `
        SELECT DISTINCT ON (ri.id)
            ri.order_item_id AS "ItemOrderID",
            oi.product_name_at_purchase AS "ProductName",
            oi.sku_at_purchase AS "SKU",
			pvi.url,
            ri.quantity AS "Quantity",
            ri.order_item_price AS "OrderItemPrice",
            ri.status AS "Status",
            ri.reason AS "Reason",
            ri.requested_at AS "RequestedAt",
            ri.approved_at AS "ApprovedAt"
        FROM return_items ri
        JOIN order_items oi ON ri.order_item_id = oi.id
		LEFT JOIN product_variants pv ON oi.product_variant_id = pv.id
		LEFT JOIN product_variant_images pvi ON pv.id = pvi.product_variant_id
        WHERE ri.return_id = $1
        ORDER BY ri.id`

	rows, err := r.DB.Query(ctx, itemsQuery, returnID)
	if err != nil {
		return nil, fmt.Errorf("failed to query return items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item domain.ReturnItem
		if err := rows.Scan(
			&item.ItemOrderID,
			&item.ProductName,
			&item.SKU,
			&item.ImageURL,
			&item.Quantity,
			&item.OrderItemPrice,
			&item.Status,
			&item.Reason,
			&item.RequestedAt,
			&item.ApprovedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan return item: %w", err)
		}
		result.Items = append(result.Items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating return items: %w", err)
	}

	return &result, nil
}

func (r OrderRepository) UpdateReturnRequestStatus(ctx context.Context, req *domain.UpdateReturnRefundRequest) (err error) {

	tx, err := r.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	// rollback only if an error occurs
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// lock the return request row to prevent concurrent updates
	var currentStatus string
	var orderID int64

	query := `
		SELECT order_id, status
		FROM order_returns
		WHERE id = $1
		FOR UPDATE
	`
	if err = tx.QueryRow(ctx, query, req.ReturnID).Scan(&orderID, &currentStatus); err != nil {
		return err
	}

	// enforce valid state transitions
	if currentStatus == "rejected" || currentStatus == "returned" {
		return apperror.ErrReturnAlreadyProcessed
	}

	// update return request status
	query = `
		UPDATE order_returns
		SET status = $1
		WHERE id = $2
	`
	if _, err = tx.Exec(ctx, query, req.Status, req.ReturnID); err != nil {
		return err
	}

	// update return items status
	if req.Status == "approved" {
		query = `
			UPDATE return_items
			SET status = 'approved',
			    approved_at = now()
			WHERE return_id = $1
			  AND status = 'requested'
		`
	} else {
		query = `
			UPDATE return_items
			SET status = 'rejected'
			WHERE return_id = $1
			  AND status = 'requested'
		`
	}

	if _, err = tx.Exec(ctx, query, req.ReturnID); err != nil {
		return err
	}

	// fetch affected order items
	query = `
		SELECT order_item_id
		FROM return_items
		WHERE return_id = $1
	`
	rows, err := tx.Query(ctx, query, req.ReturnID)
	if err != nil {
		return err
	}
	defer rows.Close()

	orderItemIDs := make([]int64, 0)
	for rows.Next() {
		var id int64
		if err = rows.Scan(&id); err != nil {
			return err
		}
		orderItemIDs = append(orderItemIDs, id)
	}

	// update order_items only on approval
	if req.Status == "approved" {

		query = `
			UPDATE order_items
			SET return_status = 'approved'
			WHERE id = ANY($1)
			  AND return_status IS DISTINCT FROM 'approved'
		`
		if _, err = tx.Exec(ctx, query, orderItemIDs); err != nil {
			return err
		}

		// check if all order items are now returned
		var remaining int
		query = `
			SELECT COUNT(*)
			FROM order_items
			WHERE order_id = $1
			  AND return_status IS DISTINCT FROM 'approved'
		`
		if err = tx.QueryRow(ctx, query, orderID).Scan(&remaining); err != nil {
			return err
		}

		// update order return_status only if full return
		if remaining == 0 {
			query = `
				UPDATE orders
				SET return_status = 'approved'
				WHERE id = $1
			`
			if _, err = tx.Exec(ctx, query, orderID); err != nil {
				return err
			}
		}
	}

	// commit transaction
	return tx.Commit(ctx)
}

func (r OrderRepository) ProcessReturnRefund(ctx context.Context, req *domain.UpdateReturnRefundRequest) (err error) {

	tx, err := r.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	// rollback only on error
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// lock return request and fetch required data
	var (
		currentStatus string
		userID        int64
		orderID       int64
	)

	query := `
		SELECT user_id, order_id, status
		FROM order_returns
		WHERE id = $1
		FOR UPDATE
	`
	if err = tx.QueryRow(ctx, query, req.ReturnID).Scan(&userID, &orderID, &currentStatus); err != nil {
		return err
	}

	// enforce valid refund state
	if currentStatus != "approved" {
		return apperror.ErrInvalidReturnState
	}

	req.UserID = userID
	req.RelatedOrder = orderID

	// mark return as returned
	query = `
		UPDATE order_returns
		SET status = 'returned'
		WHERE id = $1
	`
	if _, err = tx.Exec(ctx, query, req.ReturnID); err != nil {
		return err
	}

	// update return items status idempotently (only for approved items)
	query = `
		UPDATE return_items
		SET status = 'returned'
		WHERE return_id = $1
		  AND status = 'approved'
	`
	if _, err = tx.Exec(ctx, query, req.ReturnID); err != nil {
		return err
	}

	// update order items status
	query = `
		UPDATE order_items
		SET status = 'returned'
		WHERE id IN (
			SELECT order_item_id
			FROM return_items
			WHERE return_id = $1
			  AND status = 'returned'
		)
		  AND status != 'returned'
	`
	if _, err = tx.Exec(ctx, query, req.ReturnID); err != nil {
		return err
	}

	// check if all order items are returned
	var remaining int
	query = `
		SELECT COUNT(*)
		FROM order_items
		WHERE order_id = $1
		  AND status != 'returned'
	`
	if err = tx.QueryRow(ctx, query, orderID).Scan(&remaining); err != nil {
		return err
	}

	// update order return_status only if full return
	if remaining == 0 {
		query = `
			UPDATE orders
			SET return_status = 'returned'
			WHERE id = $1
		`
		if _, err = tx.Exec(ctx, query, orderID); err != nil {
			return err
		}
	}

	var (
		walletID      int64
		balanceBefore float64
	)

	query = `
		SELECT id, balance
		FROM wallets
		WHERE user_id = $1
		FOR UPDATE
	`
	if err = tx.QueryRow(ctx, query, userID).Scan(&walletID, &balanceBefore); err != nil {
		return err
	}

	balanceAfter := balanceBefore + req.RefundAmount

	// update wallet balance
	query = `
		UPDATE wallets
		SET balance = $1
		WHERE id = $2
	`
	if _, err = tx.Exec(ctx, query, balanceAfter, walletID); err != nil {
		return err
	}

	// record wallet transaction
	query = `
		INSERT INTO wallet_transactions (
			wallet_id,
			amount,
			transaction_type,
			related_order,
			remarks,
			balance_before,
			balance_after
		)
		VALUES ($1, $2, 'refund', $3, $4, $5, $6)
	`
	if _, err = tx.Exec(
		ctx,
		query,
		walletID,
		req.RefundAmount,
		orderID,
		req.Remarks,
		balanceBefore,
		balanceAfter,
	); err != nil {
		return err
	}

	// Restock approved items
	restockQuery := `
		UPDATE product_variants pv
		SET stock = pv.stock + ri.quantity
		FROM return_items ri
		JOIN order_items oi ON ri.order_item_id = oi.id
		WHERE ri.return_id = $1
		  AND oi.product_variant_id = pv.id
		  AND ri.status = 'returned'
	`
	cmdTag, err := tx.Exec(ctx, restockQuery, req.ReturnID)
	if err != nil {
		return fmt.Errorf("failed to restock items: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("no items to restock")
	}
	fmt.Println("Restocked items:", cmdTag.RowsAffected())

	//

	// commit transaction
	return tx.Commit(ctx)
}

// shipment
func (r OrderRepository) ShipOrder(ctx context.Context, orderID string, shipmentData *domain.ShipmentData) error {

	query := `UPDATE shipments SET carrier = $1, tracking_number = $2, status = $3,shipped_at = $4,updated_at = now()
	 WHERE order_id = (SELECT id FROM orders WHERE public_order_id = $5) RETURNING created_at,updated_at`
	err := r.DB.QueryRow(ctx, query, shipmentData.Carrier, shipmentData.TrackingID, "shipped", shipmentData.ShippedAt, orderID).Scan(&shipmentData.CreatedAt, &shipmentData.UpdatedAt)
	if err != nil {
		return err
	}

	return nil
}

// deliver order

func (r OrderRepository) DeliverOrder(ctx context.Context, orderID string) error {

	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	// update shipment status
	query := `UPDATE shipments SET status = $1,delivered_at = now(),updated_at = now()
	 WHERE order_id = (SELECT id FROM orders WHERE public_order_id = $2)`
	_, err = tx.Exec(ctx, query, "delivered", orderID)
	if err != nil {
		return err
	}

	// update payment status
	query = `UPDATE payments SET status = $1,updated_at = now()
	 WHERE order_id = (SELECT id FROM orders WHERE public_order_id = $2)`
	_, err = tx.Exec(ctx, query, "paid", orderID)
	if err != nil {
		return err
	}

	// update order status
	query = `UPDATE orders SET status = $1,updated_at = now()
	 WHERE public_order_id = $2`
	_, err = tx.Exec(ctx, query, "delivered", orderID)
	if err != nil {
		return err
	}

	tx.Commit(ctx)

	return nil
}

func (r OrderRepository) ListAllOrders(ctx context.Context, filter *domain.OrderFilter) ([]domain.OrderBaseResponse, error) {

	query := `
	SELECT 
		o.public_order_id, 
		o.total_amount, 
		o.tax_amount, 
		o.status,
		o.estimated_delivery_date,
		p.currency,
		o.shipping_address_id,
		s.type AS delivery_type,
		s.status AS shipment_status,
		p.status AS payment_status,
		a.id, a.label, a.address_line, a.address_line_2,
		a.city, a.district, a.state, a.pincode, a.country
		FROM orders o
	LEFT JOIN shipments s ON o.id = s.order_id
	LEFT JOIN payments p ON o.id = p.order_id
	LEFT JOIN user_addresses a ON o.shipping_address_id = a.id
	WHERE 1=1
	`

	var args []interface{}
	argIndex := 1

	// apply filters
	if filter != nil {

		if filter.OrderStatus != nil {
			query += fmt.Sprintf(" AND o.status = $%d", argIndex)
			args = append(args, filter.OrderStatus)
			argIndex++
		}

		if filter.DeliveryType != nil {
			query += fmt.Sprintf(" AND s.type = $%d", argIndex)
			args = append(args, filter.DeliveryType)
			argIndex++
		}

		if filter.ShipmentStatus != nil {
			query += fmt.Sprintf(" AND s.status = $%d", argIndex)
			args = append(args, filter.ShipmentStatus)
			argIndex++
		}

		if filter.PriceFrom != nil {
			query += fmt.Sprintf(" AND o.total_amount >= $%d", argIndex)
			args = append(args, filter.PriceFrom)
			argIndex++
		}

		if filter.PriceTo != nil {
			query += fmt.Sprintf(" AND o.total_amount <= $%d", argIndex)
			args = append(args, filter.PriceTo)
			argIndex++
		}

		if filter.CreatedAtFrom != nil {
			query += fmt.Sprintf(" AND o.created_at >= $%d", argIndex)
			args = append(args, filter.CreatedAtFrom)
			argIndex++
		}

		if filter.CreatedAtTo != nil {
			query += fmt.Sprintf(" AND o.created_at <= $%d", argIndex)
			args = append(args, filter.CreatedAtTo)
			argIndex++
		}
	}

	// sorting
	orderBy := "created_at"
	sort := "DESC"

	if filter != nil {
		if filter.OrderBy != nil {
			orderBy = *filter.OrderBy
		}
		if filter.Sort != nil {
			sort = strings.ToUpper(*filter.Sort)
		}
	}

	if orderBy == "price" {
		orderBy = "o.total_amount"
	}
	query += fmt.Sprintf(" ORDER BY o.%s %s", orderBy, sort)

	// pagination
	limit := 20
	page := 1

	if filter != nil {
		if filter.Limit > 0 {
			limit = filter.Limit
		}
		if filter.Page > 0 {
			page = filter.Page
		}
	}

	offset := (page - 1) * limit

	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	//  get orders
	rows, err := r.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// scan results
	var orders []domain.OrderBaseResponse

	for rows.Next() {
		var order domain.OrderBaseResponse
		var address domain.UserAddress

		err = rows.Scan(
			&order.OrderID,
			&order.TotalAmount,
			&order.TaxAmount,
			&order.Status,
			&order.EstimatedDeliveryDate,
			&order.Currency,
			&order.ShippingAddressID,
			&order.DeliveryType,
			&order.ShipmentStatus,
			&order.PaymentStatus,
			&address.ID,
			&address.Label,
			&address.AddressLine,
			&address.AddressLine2,
			&address.City,
			&address.District,
			&address.State,
			&address.Pincode,
			&address.Country,
		)
		if err != nil {
			return nil, err
		}

		order.Subtotal = order.TotalAmount - order.TaxAmount
		order.ShippingAddress = &address
		orders = append(orders, order)
	}

	return orders, nil
}
