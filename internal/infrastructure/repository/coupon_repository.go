package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
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
	err := repo.db.QueryRow(ctx, query, req.CouponCode, req.Description, req.DiscountType,
		req.DiscountValue, req.MinOrderAmount, req.MaxDiscountAmount,
		req.ValidFrom, req.ValidTo).Scan(
		&coupon.ID, &coupon.CouponCode, &coupon.Description,
		&coupon.DiscountType, &coupon.DiscountValue, &coupon.MinOrderAmount,
		&coupon.MaxDiscountAmount, &coupon.ValidFrom, &coupon.ValidTo, &coupon.IsActive,
		&coupon.CreatedAt, &coupon.UpdatedAt)
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

func (repo *CouponRepository) ListAllCoupons(ctx context.Context, filter *domain.ListCouponsFilter) ([]domain.CouponResponse, error) {

	query := `SELECT 
			id,
			code,
			description,
			discount_type,
			discount_value,
			min_order_amount,
			max_discount_amount,
			is_active,
			valid_from,
			valid_to,
			created_at,
			updated_at
		FROM coupons
		WHERE 1=1`

	args := []any{}
	argPos := 1

	if filter == nil {
		query += `
			AND is_active = true
			AND valid_from <= NOW()
			AND valid_to >= NOW()
			`
	} else {
		if filter.CouponCode != "" {
			query += fmt.Sprintf(" AND code LIKE $%d", argPos)
			args = append(args, "%"+filter.CouponCode+"%")
			argPos++
		}

		if filter.DiscountType != "" {
			query += fmt.Sprintf(" AND discount_type = $%d", argPos)
			args = append(args, filter.DiscountType)
			argPos++
		}

		if filter.ValidFrom != nil {
			query += fmt.Sprintf(" AND valid_from >= $%d", argPos)
			args = append(args, *filter.ValidFrom)
			argPos++
		}

		if filter.ValidTo != nil {
			query += fmt.Sprintf(" AND valid_to <= $%d", argPos)
			args = append(args, *filter.ValidTo)
			argPos++
		}

		if filter.IsActive != nil {
			query += fmt.Sprintf(" AND is_active = $%d", argPos)
			args = append(args, *filter.IsActive)
			argPos++
		}
	}

	query += " ORDER BY created_at DESC"

	rows, err := repo.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var coupons []domain.CouponResponse
	for rows.Next() {
		var coupon domain.CouponResponse
		err := rows.Scan(
			&coupon.ID, &coupon.CouponCode, &coupon.Description,
			&coupon.DiscountType, &coupon.DiscountValue, &coupon.MinOrderAmount,
			&coupon.MaxDiscountAmount, &coupon.IsActive, &coupon.ValidFrom, &coupon.ValidTo,
			&coupon.CreatedAt, &coupon.UpdatedAt)
		if err != nil {
			return nil, err
		}
		coupons = append(coupons, coupon)
	}

	return coupons, nil

}

func (repo *CouponRepository) FetchCoupon(ctx context.Context, couponCode string) (*domain.CouponResponse, error) {
	query := `SELECT 
		id,
		code,
		description,
		discount_type,
		discount_value,
		min_order_amount,
		max_discount_amount,
		is_active,
		valid_from,
		valid_to,
		created_at,
		updated_at
	FROM coupons
	WHERE code = $1`

	var coupon domain.CouponResponse
	err := repo.db.QueryRow(ctx, query, couponCode).Scan(
		&coupon.ID, &coupon.CouponCode, &coupon.Description,
		&coupon.DiscountType, &coupon.DiscountValue, &coupon.MinOrderAmount,
		&coupon.MaxDiscountAmount, &coupon.IsActive, &coupon.ValidFrom, &coupon.ValidTo,
		&coupon.CreatedAt, &coupon.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrCouponNotFound
		}
		return nil, err
	}

	return &coupon, nil
}
