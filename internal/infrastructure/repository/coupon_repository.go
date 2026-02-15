package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tharunn0/E-Commerce-Go/internal/apperror"
	"github.com/tharunn0/E-Commerce-Go/internal/domain"
)

type CouponRepository struct {
	db *pgxpool.Pool
}

func NewCouponRepository(db *pgxpool.Pool) *CouponRepository {
	return &CouponRepository{db: db}
}

func (repo *CouponRepository) CreateCoupon(ctx context.Context, req *domain.CreateCouponRequest) (*domain.CouponResponse, error) {

	query := `INSERT INTO coupons 
				(code, description, discount_type, discount_value, min_order_amount, max_discount_amount, valid_from, valid_to) 
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
				RETURNING id, code, description, discount_type, discount_value, min_order_amount, max_discount_amount, valid_from, valid_to, is_active, created_at, updated_at`

	var coupon domain.CouponResponse
	err := repo.db.QueryRow(ctx, query, req.CouponCode, req.Description, req.DiscountType, req.DiscountValue, req.MinOrderAmount, req.MaxDiscountAmount, req.ValidFrom, req.ValidTo).Scan(&coupon.ID, &coupon.CouponCode, &coupon.Description, &coupon.DiscountType, &coupon.DiscountValue, &coupon.MinOrderAmount, &coupon.MaxDiscountAmount, &coupon.ValidFrom, &coupon.ValidTo, &coupon.IsActive, &coupon.CreatedAt, &coupon.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			// 23505 = unique_violation
			if pgErr.Code == "23505" {
				return nil, apperror.ErrCouponCodeAlreadyExists
			}
		}
		return nil, err
	}

	return &coupon, nil
}
