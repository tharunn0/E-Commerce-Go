package auth

import (
	"context"
	"time"
)

type EmailSender interface {
	SendMail(ctx context.Context, templatePath, to, subject string, data interface{}) error
}

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
