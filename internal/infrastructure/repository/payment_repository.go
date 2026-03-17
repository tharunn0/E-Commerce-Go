package repository

import (
	"context"
	"log"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tharunn0/E-Commerce-Go/internal/apperror"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/payment"
)

type PaymentRepository struct {
	DB *pgxpool.Pool
}

func NewPaymentRepository(db *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{DB: db}
}

func (repo *PaymentRepository) CreatePayment(ctx context.Context, payment *payment.Payment) error {

	totalAmt := float64(payment.Amount) / 100
	status := strings.ToLower(string(payment.Status))

	query := `INSERT INTO payments (order_id, user_id, amount, currency, provider, provider_order_id, status)
	SELECT id, $2, $3, $4, $5, $6,$7 from orders where public_order_id = $1`
	_, err := repo.DB.Exec(ctx, query, payment.OrderID, payment.UserID, totalAmt, payment.Currency,
		payment.Provider, payment.GatewayRef, status)
	if err != nil {
		return err
	}
	return nil

}

func (repo *PaymentRepository) UpdatePaymentStatus(ctx context.Context, orderID string, status payment.PaymentStatus) error {

	var updatedPaymentStatus string

	if status == payment.PaymentStatusCompleted {
		updatedPaymentStatus = "paid"
	} else {
		updatedPaymentStatus = "failed"
	}

	query := `UPDATE payments
	SET status = $1
	WHERE provider_order_id = $2`
	cmd, err := repo.DB.Exec(ctx, query, updatedPaymentStatus, orderID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return apperror.ErrPaymentNotFound
	}
	return nil
}

func (repo *PaymentRepository) UpdateOrderPaymentStatus(ctx context.Context, publicOrderID string, paymentData *payment.WebhookPayment) error {

	tx, err := repo.DB.Begin(ctx)
	if err != nil {
		log.Println("failed to begin transaction", err)
		return err
	}
	defer tx.Rollback(ctx)

	var orderID int64

	query := `SELECT id FROM orders WHERE public_order_id = $1`
	err = tx.QueryRow(ctx, query, publicOrderID).Scan(&orderID)
	if err != nil {
		log.Println("failed to get order id", err)
		return err
	}

	log.Println("order id", orderID)

	// update order status
	query = `UPDATE orders
	SET status = $1
	WHERE status = 'pending' and public_order_id = $2`
	_, err = tx.Exec(ctx, query, "confirmed", publicOrderID)
	if err != nil {
		log.Println("failed to update order status", err)
		return err
	}

	// update order item status

	query = `UPDATE order_items
	SET status = $1
	WHERE status = 'pending' and order_id = $2`
	_, err = tx.Exec(ctx, query, "confirmed", orderID)
	if err != nil {
		log.Println("failed to update order item status", err)
		return err
	}

	// update payment status

	query = `UPDATE payments
	SET collected_amount = $1, status = $2, provider_order_id = $3, paid_at = now()
	WHERE order_id = $4`
	_, err = tx.Exec(ctx, query, paymentData.Amount/100, "paid", paymentData.ID, orderID)
	if err != nil {
		log.Println("failed to update payment status", err)
		return err
	}

	// commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		log.Println("failed to commit transaction", err)
		return err
	}
	return nil
}
