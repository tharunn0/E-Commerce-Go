package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tharunn0/E-Commerce-Go/internal/domain"
)

type AuthRepository struct {
	DB *pgxpool.Pool
}

func NewAuthRepository(db *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{DB: db}
}

func (r *AuthRepository) InsertOTP(ctx context.Context, email, otp string, expiresAt time.Time) error {

	query := `INSERT INTO otps (email, otp, expiry_at, created_at)
	VALUES ($1, $2, $3, now())`
	cmdTag, err := r.DB.Exec(ctx, query, email, otp, expiresAt)

	fmt.Println("cmdTag", cmdTag, email, otp, expiresAt, err)

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("failed to insert OTP")
	}

	return err
}

func (r *AuthRepository) GetLatestOTP(email string) (*domain.OTP, error) {

	var otp domain.OTP
	query := `SELECT id, email, otp, expiry_at, is_used, created_at
              FROM otps WHERE email = $1 ORDER BY created_at DESC LIMIT 1`
	err := r.DB.QueryRow(context.Background(), query, email).Scan(
		&otp.ID,
		&otp.Email,
		&otp.OTP,
		&otp.ExpiresAt,
		&otp.Verified,
		&otp.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &otp, nil
}

func (r *AuthRepository) MarkVerified(id int64) error {
	_, err := r.DB.Exec(context.Background(), "UPDATE otps SET is_used = true WHERE id = $1", id)
	return err
}

func (r *AuthRepository) UpdateOTP(email, otp string, expiresAt time.Time) error {
	query := `UPDATE otps 
		SET otp = $1, expires_at = $2, created_at = now() 
		WHERE email = $3`

	cmdTag, err := r.DB.Exec(context.Background(), query, otp, expiresAt, email)
	if err != nil {
		return fmt.Errorf("failed to update OTP: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("no OTP record found for email: %s", email)
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
