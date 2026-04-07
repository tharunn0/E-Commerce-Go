package auth

import (
	"time"
)

const (
	PasswordResetTokenPrefix     = "auth:password-reset:"
	EmailVerificationTokenPrefix = "auth:email-verification:"
	EmailResetTokenPrefix        = "auth:email-reset:"
)

const (
	PasswordResetTemplate     string = "password_reset.html"
	EmailVerificationTemplate string = "email_verification.html"
	EmailResetTemplate        string = "email_reset.html"
)

type PasswordResetRequest struct {
	Email string `json:"email" validate:"required,email"`
}
type PasswordResetData struct {
	Email       string `json:"email" validate:"required,email"`
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"newPassword" validate:"required,min=8"`
}
type PasswordResetToken struct {
	Token    string    `json:"token" validate:"required"`
	ExpiryAt time.Time `json:"expiryAt" validate:"required"`
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
