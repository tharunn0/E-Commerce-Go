package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/tharunn0/E-Commerce-Go/internal/apperror"
	"github.com/tharunn0/E-Commerce-Go/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type AuthRepository struct {
	DB    *pgxpool.Pool
	Redis *redis.Client
}

func NewAuthRepository(db *pgxpool.Pool, redis *redis.Client) *AuthRepository {
	return &AuthRepository{DB: db, Redis: redis}
}

func (r *AuthRepository) InsertOTP(ctx context.Context, otpdata *domain.OTP) error {

	bytes, err := json.Marshal(otpdata)
	if err != nil {
		return fmt.Errorf("failed to marshal OTP data: %w", err)
	}

	key := "auth:otp:" + otpdata.Email

	err = r.Redis.Set(ctx, key, bytes, time.Until(otpdata.ExpiresAt)).Err()
	if err != nil {
		return fmt.Errorf("failed to set OTP in Redis: %w", err)
	}

	return nil
}

func (r *AuthRepository) GetLatestOTP(email string) (*domain.OTP, error) {

	key := "auth:otp:" + email

	bytes, err := r.Redis.Get(context.Background(), key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, apperror.ErrOTPExpired
		}
		return nil, fmt.Errorf("failed to get OTP from Redis: %w", err)
	}

	var otpdata domain.OTP
	err = json.Unmarshal(bytes, &otpdata)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal OTP data: %w", err)
	}

	return &otpdata, nil
}

func (r *AuthRepository) MarkVerified(id int64) error {
	_, err := r.DB.Exec(context.Background(), "UPDATE otps SET is_used = true WHERE id = $1", id)
	return err
}

func (r *AuthRepository) UpdateOTP(email, otp string, expiresAt time.Time) error {
	// query := `UPDATE otps
	// 	SET otp = $1, expires_at = $2, created_at = now()
	// 	WHERE email = $3`

	// cmdTag, err := r.DB.Exec(context.Background(), query, otp, expiresAt, email)
	// if err != nil {
	// 	return fmt.Errorf("failed to update OTP: %w", err)
	// }

	// if cmdTag.RowsAffected() == 0 {
	// 	return fmt.Errorf("no OTP record found for email: %s", email)
	// }

	// return nil

	err := r.Redis.Set(context.Background(), email, otp, expiresAt.Sub(time.Now())).Err()
	if err != nil {
		return fmt.Errorf("failed to update OTP: %w", err)
	}

	return nil
}

func (r *AuthRepository) MarkUserVerified(email string) error {
	cmdTag, err := r.DB.Exec(context.Background(), "UPDATE users SET is_verified = true WHERE email = $1", email)
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("no user record found for email: %s", email)
	}
	return err
}
func (repo *AuthRepository) PasswordReset(ctx context.Context, req *domain.PasswordResetData) error {
	query := `UPDATE users SET password = $1 WHERE email = $2`
	cmdTag, err := repo.DB.Exec(ctx, query, req.NewPassword, req.Email)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("FAILED_TO_RESET_PASSWORD: NO USER FOUND")
	}
	return nil
}

func (repo *AuthRepository) SetPasswordResetToken(ctx context.Context, req *domain.PasswordResetToken) error {
	query := `INSERT INTO password_reset_tokens (token,expiry_at) VALUES ($1,$2)`
	cmdTag, err := repo.DB.Exec(ctx, query, req.Token, req.ExpiryAt)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("FAILED_TO_SET_PASSWORD_RESET_TOKEN")
	}
	return nil
}

func (repo *AuthRepository) GetPasswordResetToken(ctx context.Context, token string) (*domain.PasswordResetToken, error) {
	query := `SELECT token,expiry_at FROM password_reset_tokens WHERE token = $1 AND is_used = FALSE`
	rows := repo.DB.QueryRow(ctx, query, token)
	var passwordResetToken domain.PasswordResetToken
	err := rows.Scan(&passwordResetToken.Token, &passwordResetToken.ExpiryAt)

	if err == pgx.ErrNoRows {
		return nil, errors.New("INVALID_OR_EXPIRED_TOKEN")
	}

	if err != nil {
		return nil, err
	}

	return &passwordResetToken, nil
}

func (repo *AuthRepository) InvalidatePasswordResetToken(ctx context.Context, token string) error {
	query := `UPDATE password_reset_tokens SET is_used = TRUE WHERE token = $1`
	cmdTag, err := repo.DB.Exec(ctx, query, token)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("FAILED_TO_INVALIDATE_PASSWORD_RESET_TOKEN")
	}
	return nil
}
