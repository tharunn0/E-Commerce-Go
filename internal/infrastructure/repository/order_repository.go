package repository

import (
	// "context"

	"context"
	"fmt"
	"strings"

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

	// begin transaction
	tx, err := r.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		fmt.Println("error beginning transaction", err)
		return err
	}

	// ensure rollback on error
	defer tx.Rollback(ctx)

	var internalOrderID int64
	// insert into orders
	query := `INSERT INTO orders (user_id, public_order_id, total_amount, tax_amount, shipping_address_id, billing_address_id,estimated_delivery_date,status)
	VALUES ($1, $2, $3, $4, $5, $6,$7,$8) RETURNING id`
	err = tx.QueryRow(ctx, query, data.UserID, data.OrderID, data.TotalAmount, data.TaxAmount, data.ShippingAddressID, data.BillingAddressID, data.EstimatedDeliveryDate, "confirmed").Scan(&internalOrderID)
	if err != nil {
		fmt.Println("error inserting order", err)
		return err
	}

	// insert into order_items
	for _, item := range data.Items {
		query = `INSERT INTO order_items (order_id, product_variant_id, sku_at_purchase, product_name_at_purchase, quantity, unit_price, total_price,status)
	SELECT $1, $2, pv.sku, p.name, $3, $4, $5, $6 FROM product_variants pv
	JOIN products p ON pv.product_id = p.id WHERE pv.id = $2`
		_, err = tx.Exec(ctx, query, internalOrderID, item.ProductVariantID, item.Quantity, item.UnitPrice, item.TotalPrice, "purchased")
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
	// commit transaction
	return tx.Commit(ctx)
}

func (r OrderRepository) GetUserOrders(ctx context.Context, userID int64) ([]domain.OrderBaseResponse, error) {

	query := `SELECT o.public_order_id, o.total_amount, o.tax_amount, o.status,o.estimated_delivery_date,p.currency,o.shipping_address_id,
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
		err = rows.Scan(&order.OrderID, &order.TotalAmount, &order.TaxAmount, &order.Status, &order.EstimatedDeliveryDate, &order.Currency, &order.ShippingAddressID,
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

func (r OrderRepository) GetUserOrderByID(ctx context.Context, orderID string) (*domain.OrderResponse, error) {
	query := `SELECT o.public_order_id, o.total_amount, o.tax_amount, o.status,o.estimated_delivery_date,p.currency,o.shipping_address_id,o.billing_address_id,
		s.type as delivery_type,s.status as shipment_status,p.status as payment_status,p.provider,o.created_at,o.updated_at
		FROM orders o 
		LEFT JOIN shipments s ON o.id = s.order_id
		LEFT JOIN payments p ON o.id = p.order_id
		WHERE o.public_order_id = $1`
	var order domain.OrderResponse
	row := r.DB.QueryRow(ctx, query, orderID)
	err := row.Scan(&order.OrderID, &order.TotalAmount, &order.TaxAmount, &order.Status, &order.EstimatedDeliveryDate, &order.Currency,
		&order.ShippingAddressID, &order.BillingAddressID,
		&order.DeliveryType, &order.ShipmentStatus, &order.PaymentStatus, &order.PaymentMethod,
		&order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			fmt.Println("order not found", orderID)
			return nil, apperror.ErrOrderNotFound
		}
		return nil, err
	}

	items := []domain.OrderItem{}
	query = `SELECT pv.id, p.name, pv.sku, oi.quantity, oi.unit_price, oi.total_price
		FROM order_items oi
		LEFT JOIN product_variants pv ON oi.product_variant_id = pv.id
		LEFT JOIN products p ON pv.product_id = p.id
		WHERE oi.order_id = (SELECT id FROM orders WHERE public_order_id = $1)`
	rows, err := r.DB.Query(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item domain.OrderItem
		err = rows.Scan(&item.ProductVariantID, &item.ProductName, &item.SKU, &item.Quantity, &item.UnitPrice, &item.TotalPrice)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	order.Items = items

	order.Subtotal = order.TotalAmount - order.TaxAmount
	return &order, nil
}

func (r OrderRepository) CancelOrderItem(ctx context.Context, orderID string, variantID int64) error {
	// update order item status
	query := `UPDATE order_items SET status = $1 WHERE order_id = (SELECT id FROM orders 
	WHERE public_order_id = $2) AND product_variant_id = $3`
	_, err := r.DB.Exec(ctx, query, "cancelled", orderID, variantID)
	if err != nil {
		return err
	}

	// update order prices
	query = `UPDATE orders SET total_amount = (SELECT SUM(total_price) FROM order_items
	 WHERE order_id = (SELECT id FROM orders WHERE public_order_id = $1)) WHERE public_order_id = $1`
	_, err = r.DB.Exec(ctx, query, orderID)
	if err != nil {
		return err
	}

	// update order status
	query = `UPDATE orders SET status = $1 WHERE public_order_id = $2`
	_, err = r.DB.Exec(ctx, query, "cancelled", orderID)
	if err != nil {
		return err
	}
	return nil
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
