package repository

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tharunn0/E-Commerce-Go/internal/apperror"
	"github.com/tharunn0/E-Commerce-Go/internal/domain"
)

type PaymentRepository struct {
	DB *pgxpool.Pool
}

func NewPaymentRepository(db *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{DB: db}
}

func (p *PaymentRepository) CreatePayment(ctx context.Context, payment *domain.Payment) error {

	totalAmt := float64(payment.Amount) / 100
	status := strings.ToLower(string(payment.Status))

	query := `INSERT INTO payments (order_id, user_id, amount, currency, provider, provider_order_id, status)
	SELECT id, $2, $3, $4, $5, $6,$7 from orders where public_order_id = $1`
	_, err := p.DB.Exec(ctx, query, payment.OrderID, payment.UserID, totalAmt, payment.Currency,
		payment.Provider, payment.GatewayRef, status)
	if err != nil {
		return err
	}
	return nil

}

func (p *PaymentRepository) UpdatePaymentStatus(ctx context.Context, orderID string, status domain.PaymentStatus) error {

	var updatedPaymentStatus string

	if status == domain.PaymentStatusCompleted {
		updatedPaymentStatus = "paid"
	} else {
		updatedPaymentStatus = "failed"
	}

	query := `UPDATE payments 
	SET status = $1 
	WHERE provider_order_id = $2`
	cmd, err := p.DB.Exec(ctx, query, updatedPaymentStatus, orderID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return apperror.ErrPaymentNotFound
	}
	return nil
}
