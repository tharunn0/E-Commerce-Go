package repository

import (
	// "context"

	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	var internalOrderID int64

	// insert into orders
	query := `INSERT INTO orders (user_id, public_order_id, total_amount, tax_amount, shipping_address_id, billing_address_id,estimated_delivery_date)
	VALUES ($1, $2, $3, $4, $5, $6,$7) RETURNING id`
	err = tx.QueryRow(ctx, query, data.UserID, data.OrderID, data.TotalAmount, data.TaxAmount, data.ShippingAddressID, data.BillingAddressID, data.EstimatedDeliveryDate).Scan(&internalOrderID)
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

	// insert into payments
	query = `INSERT INTO payments (order_id, user_id, amount, currency,provider, status)
	VALUES ($1, $2, $3, $4, $5, $6)`
	_, err = tx.Exec(ctx, query, internalOrderID, data.UserID, data.TotalAmount, "INR", "COD", "pending")
	if err != nil {
		fmt.Println("error inserting payment", err)
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
	// commit transaction
	return tx.Commit(ctx)
}

func (r OrderRepository) GetUserOrders(ctx context.Context, userID int64) ([]domain.OrderBaseResponse, error) {
	// OrderID     string  `json:"order_id"`
	// Subtotal    float64 `json:"subtotal"`
	// TaxAmount   float64 `json:"tax_amount"`
	// TotalAmount float64 `json:"total_amount"`
	// Currency    string  `json:"currency"`

	// ShippingAddressID int64 `json:"shipping_address_id"`

	// DeliveryType          DeliveryType `json:"delivery_type"`
	// EstimatedDeliveryTime string       `json:"estimated_delivery_time"`

	// Status         OrderStatus `json:"order_status"`
	// ShipmentStatus string      `json:"shipment_status"`
	// PaymentStatus  string      `json:"payment_status"`
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
