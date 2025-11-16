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
	SetEmailVerificationToken(ctx context.Context, req *EmailVerificationTokenData) error
	GetEmailVerificationToken(ctx context.Context, token string) (*EmailVerificationTokenData, error)
	UpdateEmail(ctx context.Context, currentEmail string, req *EmailVerificationTokenData) error

	// refresh token
	SetRefreshToken(ctx context.Context, req *RefreshToken) error
	GetRefreshToken(ctx context.Context, token string) (*RefreshToken, *RefreshTokenUserData, error)
	RevokeRefreshToken(ctx context.Context, token string) error
}

type RefreshToken struct {
	UserID   int64     `json:"user_id" `
	Token    string    `form:"token" `
	Revoked  bool      `json:"revoked" `
	ExpiryAt time.Time `json:"expiryAt" `
}

type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenUserData struct {
	UserID     int64  `json:"user_id" `
	Email      string `json:"email" `
	Role       string `json:"role" `
	IsVerified bool   `json:"isVerified" `
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

type EmailVerificationTokenData struct {
	Email    string
	Token    string
	ExpiryAt time.Time
}
