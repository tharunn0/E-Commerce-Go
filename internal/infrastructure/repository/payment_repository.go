package repository

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tharunn0/E-Commerce-Go/internal/domain"
)

type PaymentRepository struct {
	DB *pgxpool.Pool
}

func NewPaymentRepository(db *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{DB: db}
}

func (p *PaymentRepository) CreatePayment(ctx context.Context, payment *domain.Payment) error {

	//  order_id
	//  user_id
	//  amount
	//  collected_amount
	//  refund_amount
	//  currency
	//  provider
	//  provider_payment_id
	//  provider_payment_method
	//  status
	//  failure_reason
	//  idempotency_key
	//  paid_at
	//  refunded_at

	// insert into payments

	totalAmt := float64(payment.Amount) / 100
	status := strings.ToLower(string(payment.Status))

	query := `INSERT INTO payments (order_id, user_id, amount, currency, provider, provider_payment_id, status)
	SELECT id, $2, $3, $4, $5, $6,$7 from orders where public_order_id = $1`
	_, err := p.DB.Exec(ctx, query, payment.OrderID, payment.UserID, totalAmt, payment.Currency, payment.Provider, payment.GatewayRef, status)
	if err != nil {
		return err
	}
	return nil

}
