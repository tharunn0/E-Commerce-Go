package service

import (
	"context"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/tharunn0/E-Commerce-Go/internal/domain"
	"github.com/tharunn0/E-Commerce-Go/internal/infrastructure/repository"
	"github.com/tharunn0/E-Commerce-Go/internal/utils"
	"github.com/tharunn0/E-Commerce-Go/pkg/mailer"
	"go.uber.org/zap"
)

type AuthService struct {
	repo   repository.AuthRepository
	sender *mailer.MailSender
	log    *zap.Logger
}

func NewAuthService(Repo repository.AuthRepository, Sender *mailer.MailSender, logger *zap.Logger) *AuthService {
	return &AuthService{
		repo:   Repo,
		sender: Sender,
		log:    logger,
	}
}

func (serv *AuthService) SendOTP(toAddr string) *domain.APIError {
	ctx := context.Background()
	exp, err := strconv.Atoi(os.Getenv("OTP_EXPIRY_TIME"))
	if err != nil {
		return &domain.APIError{
			Code:    "OTP_EXPIRY_TIME_CONVERSION_FAILED",
			Message: "Could not convert OTP expiry time. Please try again.",
		}
	}

	otp, err := utils.GenerateOTP()
	if err != nil {
		return &domain.APIError{
			Code:    "OTP_GENERATION_FAILED",
			Message: "Could not generate OTP. Please try again.",
		}
	}

	expiresAt := time.Now().Add(time.Duration(exp) * time.Minute)
	if err := serv.repo.InsertOTP(ctx, toAddr, otp, expiresAt); err != nil {
		return &domain.APIError{
			Code:    "OTP_SAVE_FAILED",
			Message: "Could not save OTP. Please try again.",
		}
	}

	// body := fmt.Sprintf("Your OTP is %s. Use within 5 minutes.", otp)

	data := struct {
		OTP    string
		Expiry int
		Year   string
	}{
		OTP:    otp,
		Expiry: exp,
		Year:   strconv.Itoa(time.Now().Year()),
	}

	toAddr = strings.ToLower(toAddr)

	if err := serv.sender.SendMail(ctx, "./pkg/mailer/verification-mail-template.html", toAddr, "Email Verification", data); err != nil {
		return &domain.APIError{
			Code:    "OTP_EMAIL_FAILED",
			Message: "Could not send OTP email. Please try again.",
		}
	}

	serv.log.Info("OTP sent successfully", zap.String("email", toAddr))
	return nil
}

func (serv *AuthService) VerifyOTP(email, otp string) *domain.APIError {

	otpRecord, err := serv.repo.GetLatestOTP(email)
	if err != nil {
		serv.log.Error("OTP verification failed", zap.String("email", email), zap.Error(err))
		return &domain.APIError{
			Code:    "OTP_FETCH_FAILED",
			Message: "Something went wrong. Please try again.",
		}
	}

	if time.Now().After(otpRecord.ExpiresAt) {
		return &domain.APIError{
			Code:    "OTP_EXPIRED",
			Message: "Your OTP has expired. Please request a new one.",
		}
	}

	if otpRecord.OTP != otp {
		return &domain.APIError{
			Code:    "INVALID_OTP",
			Message: "The OTP you entered is incorrect. Please try again.",
		}
	}

	if err := serv.repo.MarkVerified(otpRecord.ID); err != nil {
		return &domain.APIError{
			Code:    "OTP_MARK_VERIFIED_FAILED",
			Message: "OTP verification successful but failed to update status. Please contact support.",
		}
	}

	if err := serv.repo.MarkUserVerified(email); err != nil {
		return &domain.APIError{
			Code:    "USER_MARK_VERIFIED_FAILED",
			Message: "OTP verification successful but failed to update user status. Please contact support.",
		}
	}

	serv.log.Info("OTP verified successfully", zap.String("email", email))
	return nil
}

func (serv *AuthService) SendPasswordResetLink(ctx context.Context, req *domain.PasswordResetRequest) *domain.APIError {

	if !utils.IsValidEmail(req.Email) {
		return &domain.APIError{
			Code:    "INVALID_EMAIL",
			Message: "Please provide a valid email address.",
		}
	}

	token, err := utils.GenerateToken(16)
	if err != nil {
		return &domain.APIError{
			Code:    "TOKEN_GENERATION_FAILED",
			Message: "Something went wrong generating a reset token.",
		}
	}

	expiryTimeStr := os.Getenv("PASSWORD_RESET_TOKEN_EXPIRY_TIME")
	expiryTime, err := strconv.Atoi(expiryTimeStr)
	if err != nil || expiryTime == 0 {
		expiryTime = 15
	}

	expiryAt := time.Now().Add(time.Duration(expiryTime) * time.Minute)

	resetToken := &domain.PasswordResetToken{
		Token:    token,
		ExpiryAt: expiryAt,
	}

	err = serv.repo.SetPasswordResetToken(ctx, resetToken)
	if err != nil {
		return &domain.APIError{
			Code:    "TOKEN_SAVE_FAILED",
			Message: "Could not initiate password reset. Please try again.",
		}
	}

	resetLink := "/reset-password?token=" + token
	emailData := struct {
		ResetLink string
		Expiry    int
		Year      string
	}{
		ResetLink: resetLink,
		Expiry:    expiryTime,
		Year:      strconv.Itoa(time.Now().Year()),
	}
	// "./pkg/mailer/verification-mail-template.html"

	if err := serv.sender.SendMail(ctx, "./pkg/mailer/reset_password_mail.html", req.Email, "Password Reset Request", emailData); err != nil {
		serv.log.Debug("Failed to send password reset email", zap.String("email", req.Email), zap.Error(err))
		return &domain.APIError{
			Code:    "SEND_MAIL_FAILED",
			Message: "Failed to send password reset email. Please try again.",
		}
	}
	serv.log.Info("Password reset link sent successfully", zap.String("email", req.Email))
	return nil
}

func (serv *AuthService) ResetPassword(ctx context.Context, data *domain.PasswordResetData) *domain.APIError {

	if !utils.IsValidPassword(data.NewPassword) {
		return &domain.APIError{
			Code:    "WEAK_PASSWORD",
			Message: "Password must contain at least 8 characters, including numbers/symbols.",
		}
	}

	resetToken, err := serv.repo.GetPasswordResetToken(ctx, data.Token)
	if err != nil || resetToken == nil {
		serv.log.Debug("Failed to get password reset token", zap.String("token", data.Token), zap.Error(err))
		return &domain.APIError{
			Code:    "INVALID_OR_EXPIRED_TOKEN",
			Message: "Password reset token is invalid or expired.",
		}
	}

	if time.Now().After(resetToken.ExpiryAt) {
		return &domain.APIError{
			Code:    "TOKEN_EXPIRED",
			Message: "Password reset token has expired.",
		}
	}

	hashedPass := utils.HashPassword(data.NewPassword)
	if hashedPass == "" {
		return &domain.APIError{
			Code:    "HASHING_FAILED",
			Message: "Password hashing failed. Please try again.",
		}
	}

	data.NewPassword = hashedPass

	err = serv.repo.PasswordReset(ctx, data)
	if err != nil {
		return &domain.APIError{
			Code:    "PASSWORD_UPDATE_FAILED",
			Message: "Failed to update password. Please try again.",
		}
	}

	err = serv.repo.InvalidatePasswordResetToken(ctx, data.Token)
	if err != nil {
	}

	serv.log.Info("password reset successful", zap.String("email", data.Email))
	return nil
}
