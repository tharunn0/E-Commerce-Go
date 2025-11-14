package domain

import (
	"context"
	"time"
)

const (
	PasswordResetTokenPrefix     = "auth:password-reset:"
	EmailVerificationTokenPrefix = "auth:email-verification:"
	EmailResetTokenPrefix        = "auth:email-reset:"
)

type AuthRepository interface {

	// user verification
	InsertOTP(ctx context.Context, otpdata *OTP) error
	GetLatestOTP(email string) (*OTP, error)
	MarkVerified(id int64) error
	UpdateOTP(email, otp string, expiresAt time.Time) error
	MarkUserVerified(email string) error

	// password reset
	PasswordReset(ctx context.Context, req *PasswordResetData) error
	SetPasswordResetToken(ctx context.Context, req *PasswordResetToken) error
	GetPasswordResetToken(ctx context.Context, token string) (*PasswordResetToken, error)

	// email verification
	SetEmailVerificationToken(ctx context.Context, req *AuthTokenData) error
	GetEmailVerificationToken(ctx context.Context, token string) (*AuthTokenData, error)
	UpdateEmail(ctx context.Context, currentEmail string, req *AuthTokenData) error
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

type UpdateEmailRequest struct {
	Email string `json:"email" validate:"required,email"`
}
type VerifyEmailRequest struct {
	Token string `form:"token" validate:"required"`
}

type AuthTokenData struct {
	Email    string
	Token    string
	ExpiryAt time.Time
}
