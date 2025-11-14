package domain

import (
	"context"
	"time"
)

type AuthRepository interface {
	InsertOTP(ctx context.Context, otpdata *OTP) error
	GetLatestOTP(email string) (*OTP, error)
	MarkVerified(id int64) error
	UpdateOTP(email, otp string, expiresAt time.Time) error
	MarkUserVerified(email string) error
	PasswordReset(ctx context.Context, req *PasswordResetData) error
	SetPasswordResetToken(ctx context.Context, req *PasswordResetToken) error
	GetPasswordResetToken(ctx context.Context, token string) (*PasswordResetToken, error)
	InvalidatePasswordResetToken(ctx context.Context, token string) error
}

type OTP struct {
	ID        int64
	Email     string `json:"email" validate:"required,email"`
	OTP       string
	ExpiresAt time.Time
	Verified  bool
	CreatedAt time.Time
}

type VerifyOTPReq struct {
	Email string `json:"email" validate:"required,email"`
	OTP   string `json:"otp" validate:"required,len=6"`
}
