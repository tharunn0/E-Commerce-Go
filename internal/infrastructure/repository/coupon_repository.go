package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tharunn0/E-Commerce-Go/internal/apperror"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/promotion"
)

type CouponRepository struct {
	db *pgxpool.Pool
}

func NewCouponRepository(db *pgxpool.Pool) *CouponRepository {
	return &CouponRepository{db: db}
}

func (repo *CouponRepository) CreateCoupon(ctx context.Context, req *promotion.CreateCouponRequest) (*promotion.CouponResponse, error) {

	query := `INSERT INTO coupons 
				(code, description, discount_type, discount_value, min_order_amount, max_discount_amount, valid_from, valid_to) 
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
				RETURNING id, code, description, discount_type, discount_value, min_order_amount, max_discount_amount, valid_from, valid_to, is_active, created_at, updated_at`

	var coupon promotion.CouponResponse
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

func (repo *CouponRepository) ListAllCoupons(ctx context.Context, filter *promotion.ListCouponsFilter) ([]promotion.CouponResponse, error) {

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

	var coupons []promotion.CouponResponse
	for rows.Next() {
		var coupon promotion.CouponResponse
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

func (repo *CouponRepository) FetchCoupon(ctx context.Context, couponCode string, userID int64) (*promotion.CouponResponse, error) {

	// fetch referral first ,if not exist move to coupons
	query := `SELECT 
		id,
		coupon_code,
		discount_amount,
		min_order_amount,
		created_at
	FROM referral_coupons
	WHERE coupon_code = $1 AND is_used = false AND user_id = $2`
	var coupon promotion.CouponResponse
	err := repo.db.QueryRow(ctx, query, couponCode, userID).Scan(
		&coupon.ID, &coupon.CouponCode, &coupon.DiscountValue, &coupon.MinOrderAmount,
		&coupon.CreatedAt)

	log.Println("Referral Coupon : ", coupon)

	coupon.DiscountType = "fixed"
	coupon.MaxDiscountAmount = coupon.DiscountValue
	coupon.IsActive = true
	coupon.ValidFrom = time.Now()
	coupon.ValidTo = time.Now().AddDate(1, 0, 0)

	if err == nil {
		return &coupon, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		log.Println("Error fetching referral coupon:", err)
		return nil, err
	}

	query = `SELECT 
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

	err = repo.db.QueryRow(ctx, query, couponCode).Scan(
		&coupon.ID, &coupon.CouponCode, &coupon.Description,
		&coupon.DiscountType, &coupon.DiscountValue, &coupon.MinOrderAmount,
		&coupon.MaxDiscountAmount, &coupon.IsActive, &coupon.ValidFrom, &coupon.ValidTo,
		&coupon.CreatedAt, &coupon.UpdatedAt)
	if err != nil {
		log.Println("Error fetching coupon:", err)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrCouponNotFound
		}
		return nil, err
	}

	return &coupon, nil
}

func (repo *CouponRepository) ListReferralRewards(ctx context.Context, userID int64) ([]promotion.CouponResponse, error) {
	query := `SELECT id,
		coupon_code,
		discount_amount,
		min_order_amount,
		created_at
	FROM referral_coupons
	WHERE user_id = $1 AND is_used = false`

	rows, err := repo.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var coupons []promotion.CouponResponse
	for rows.Next() {
		var coupon promotion.CouponResponse
		err := rows.Scan(
			&coupon.ID, &coupon.CouponCode, &coupon.DiscountValue, &coupon.MinOrderAmount,
			&coupon.CreatedAt)
		if err != nil {
			return nil, err
		}
		coupons = append(coupons, coupon)
	}

	for i := range coupons {
		coupons[i].Description = "Referral Coupon"
		coupons[i].DiscountType = "FIXED"
		coupons[i].MaxDiscountAmount = coupons[i].DiscountValue
		coupons[i].IsActive = true
		coupons[i].ValidFrom = time.Now()
		coupons[i].ValidTo = coupons[i].CreatedAt.AddDate(2, 0, 0)
	}

	return coupons, nil
}

func (repo *CouponRepository) FetchCouponByID(ctx context.Context, couponID int64) (*promotion.CouponResponse, error) {

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
	WHERE id = $1`

	var coupon promotion.CouponResponse
	err := repo.db.QueryRow(ctx, query, couponID).Scan(
		&coupon.ID, &coupon.CouponCode, &coupon.Description,
		&coupon.DiscountType, &coupon.DiscountValue, &coupon.MinOrderAmount,
		&coupon.MaxDiscountAmount, &coupon.IsActive, &coupon.ValidFrom, &coupon.ValidTo,
		&coupon.CreatedAt, &coupon.UpdatedAt)
	if err != nil {
		log.Println("Error fetching coupon:", err)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrCouponNotFound
		}
		return nil, err
	}

	return &coupon, nil
}

func (repo *CouponRepository) UpdateCoupon(ctx context.Context, req *promotion.UpdateCouponRequest) (*promotion.CouponResponse, error) {

	query := `UPDATE coupons 
		SET 
			description = COALESCE($2, description),
			discount_type = COALESCE($3, discount_type),
			discount_value = COALESCE($4, discount_value),
			min_order_amount = COALESCE($5, min_order_amount),
			max_discount_amount = COALESCE($6, max_discount_amount),
			valid_from = COALESCE($7, valid_from),
			valid_to = COALESCE($8, valid_to),
			is_active = COALESCE($9, is_active)
		WHERE id = $1
		RETURNING id, code, description, discount_type, discount_value, min_order_amount, max_discount_amount, is_active, valid_from, valid_to, created_at, updated_at`

	var coupon promotion.CouponResponse
	err := repo.db.QueryRow(ctx, query, req.ID, req.Description,
		req.DiscountType, req.DiscountValue, req.MinOrderAmount,
		req.MaxDiscountAmount, req.ValidFrom, req.ValidTo, req.IsActive).Scan(
		&coupon.ID, &coupon.CouponCode, &coupon.Description,
		&coupon.DiscountType, &coupon.DiscountValue, &coupon.MinOrderAmount,
		&coupon.MaxDiscountAmount, &coupon.IsActive, &coupon.ValidFrom, &coupon.ValidTo,
		&coupon.CreatedAt, &coupon.UpdatedAt)
	if err != nil {
		log.Println("Error updating coupon:", err)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrCouponNotFound
		}
		return nil, err
	}

	return &coupon, nil

}
