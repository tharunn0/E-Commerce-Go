package service

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/tharunn0/E-Commerce-Go/internal/apperror"
	"github.com/tharunn0/E-Commerce-Go/internal/config"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/auth"
	"github.com/tharunn0/E-Commerce-Go/internal/utils"
	"github.com/tharunn0/E-Commerce-Go/pkg/mailer"
	"go.uber.org/zap"
)

type AuthService struct {
	repo   auth.AuthRepository
	sender *mailer.MailSender
	log    *zap.Logger

	cfg *config.SecuritySettings
}

func NewAuthService(Repo auth.AuthRepository, Sender *mailer.MailSender, logger *zap.Logger, cfg *config.SecuritySettings) *AuthService {
	return &AuthService{
		repo:   Repo,
		sender: Sender,
		log:    logger,
		cfg:    cfg,
	}
}

// account verification
func (serv *AuthService) SendOTP(ctx context.Context, toAddr string) *apperror.APIError {

	otp, err := utils.GenerateOTP()
	if err != nil {
		serv.log.Error("OTP generation failed", zap.String("service", "auth-service"), zap.String("function", "sendotp"), zap.Error(err))
		return &apperror.APIError{
			Code:    "OTP_GENERATION_FAILED",
			Message: "Could not generate OTP. Please try again.",
		}
	}

	otpHash := utils.HashOTP(otp)
	exp := serv.cfg.OTPExpiryMinutes
	expiresAt := time.Now().Add(time.Duration(exp) * time.Minute)
	otpdata := &auth.OTP{
		Email:     toAddr,
		OTP:       otpHash,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}
	if err := serv.repo.InsertOTP(ctx, otpdata); err != nil {
		return &apperror.APIError{
			Code:    "OTP_SAVE_FAILED",
			Message: "Could not save OTP. Please try again.",
		}
	}

	data := struct {
		OTP             string
		ExpiryInMinutes int
		Year            string
	}{
		OTP:             otp,
		ExpiryInMinutes: exp,
		Year:            strconv.Itoa(time.Now().Year()),
	}

	toAddr = strings.ToLower(toAddr)

	if err := serv.sender.SendMail(ctx, "./pkg/mailer/verification-mail-template.html", toAddr, "Email Verification", data); err != nil {
		serv.log.Error("SEND_OTP_FAILED", zap.String("service", "auth-service"), zap.String("function", "sendotp"), zap.Error(err))
		return &apperror.APIError{
			Code:    "OTP_EMAIL_FAILED",
			Message: "Could not send OTP email. Please try again.",
		}
	}

	serv.log.Info("OTP sent successfully", zap.String("email", toAddr))
	return nil
}

func (serv *AuthService) VerifyOTP(email, otp string) *apperror.APIError {

	otpRecord, err := serv.repo.GetLatestOTP(email)
	if err != nil {
		if err == apperror.ErrOTPExpired {
			return &apperror.APIError{
				Code:    "OTP_EXPIRED",
				Message: "Your OTP has expired. Please request a new one.",
			}
		}
		serv.log.Error("OTP verification failed", zap.String("email", email), zap.Error(err))
		return &apperror.APIError{
			Code:    "OTP_FETCH_FAILED",
			Message: "Something went wrong. Please try again.",
		}
	}

	if !utils.VerifyOTP(otpRecord.OTP, otp) {
		return &apperror.APIError{
			Code:    "INVALID_OTP",
			Message: "The OTP you entered is incorrect. Please try again.",
		}
	}

	if time.Now().After(otpRecord.ExpiresAt) {
		return &apperror.APIError{
			Code:    "OTP_EXPIRED",
			Message: "Your OTP has expired. Please request a new one.",
		}
	}

	if err := serv.repo.MarkVerified(otpRecord.ID); err != nil {
		return &apperror.APIError{
			Code:    "OTP_MARK_VERIFIED_FAILED",
			Message: "OTP verification successful but failed to update status. Please contact support.",
		}
	}

	if err := serv.repo.MarkUserVerified(email); err != nil {
		return &apperror.APIError{
			Code:    "USER_MARK_VERIFIED_FAILED",
			Message: "OTP verification successful but failed to update user status. Please contact support.",
		}
	}

	serv.log.Info("OTP verified successfully", zap.String("email", email))
	return nil
}

// password reset
func (serv *AuthService) SendPasswordResetLink(ctx context.Context, req *auth.PasswordResetRequest) *apperror.APIError {

	if !utils.IsValidEmail(req.Email) {
		return &apperror.APIError{
			Code:    "INVALID_EMAIL",
			Message: "Please provide a valid email address.",
		}
	}

	token, err := utils.GenerateToken(16)
	if err != nil {
		return &apperror.APIError{
			Code:    "TOKEN_GENERATION_FAILED",
			Message: "Something went wrong generating a reset token.",
		}
	}

	expiryTime := serv.cfg.PasswordResetExpiry
	expiryAt := time.Now().Add(time.Duration(expiryTime) * time.Minute)

	resetToken := &auth.PasswordResetToken{
		Token:    token,
		ExpiryAt: expiryAt,
	}

	err = serv.repo.SetPasswordResetToken(ctx, resetToken)
	if err != nil {
		return &apperror.APIError{
			Code:    "TOKEN_SAVE_FAILED",
			Message: "Could not initiate password reset. Please try again.",
		}
	}

	resetLink := "/reset-password?token=" + token
	emailData := struct {
		ResetLink       string
		ExpiryInMinutes int
		Year            string
	}{
		ResetLink:       resetLink,
		ExpiryInMinutes: expiryTime,
		Year:            strconv.Itoa(time.Now().Year()),
	}
	// "./pkg/mailer/verification-mail-template.html"

	if err := serv.sender.SendMail(ctx, "./pkg/mailer/reset_password_mail.html", req.Email, "Password Reset Request", emailData); err != nil {
		serv.log.Debug("Failed to send password reset email", zap.String("email", req.Email), zap.Error(err))
		return &apperror.APIError{
			Code:    "SEND_MAIL_FAILED",
			Message: "Failed to send password reset email. Please try again.",
		}
	}
	serv.log.Info("Password reset link sent successfully", zap.String("email", req.Email))
	return nil
}

func (serv *AuthService) ResetPassword(ctx context.Context, data *auth.PasswordResetData) *apperror.APIError {

	if !utils.IsValidPassword(data.NewPassword) {
		return &apperror.APIError{
			Code:    "WEAK_PASSWORD",
			Message: "Password must contain at least 8 characters, including numbers/symbols.",
		}
	}

	resetToken, err := serv.repo.GetPasswordResetToken(ctx, data.Token)
	if err != nil || resetToken == nil {
		serv.log.Debug("Failed to get password reset token", zap.String("token", data.Token), zap.Error(err))
		return &apperror.APIError{
			Code:    "INVALID_OR_EXPIRED_TOKEN",
			Message: "Password reset token is invalid or expired.",
		}
	}

	if time.Now().After(resetToken.ExpiryAt) {
		return &apperror.APIError{
			Code:    "TOKEN_EXPIRED",
			Message: "Password reset token has expired.",
		}
	}

	hashedPass := utils.HashPassword(data.NewPassword)
	if hashedPass == "" {
		return &apperror.APIError{
			Code:    "HASHING_FAILED",
			Message: "Password hashing failed. Please try again.",
		}
	}

	data.NewPassword = hashedPass

	err = serv.repo.PasswordReset(ctx, data)
	if err != nil {
		return &apperror.APIError{
			Code:    "PASSWORD_UPDATE_FAILED",
			Message: "Failed to update password. Please try again.",
		}
	}

	serv.log.Info("password reset successful", zap.String("email", data.Email))
	return nil
}

// email reset
func (serv *AuthService) SendEmailChangeLink(ctx context.Context, req *auth.UpdateEmailRequest) *apperror.APIError {
	if !utils.IsValidEmail(req.Email) {
		return &apperror.APIError{
			Code:    "INVALID_EMAIL",
			Message: "Please provide a valid email address.",
		}
	}

	// generate token
	token, err := utils.GenerateToken(16)
	if err != nil {
		return &apperror.APIError{
			Code:    "TOKEN_GENERATION_FAILED",
			Message: "Something went wrong generating a token.",
		}
	}

	expiryTime := serv.cfg.EmailVerificationExpiry
	expiryAt := time.Now().Add(time.Duration(expiryTime) * time.Minute)

	// set token in redis
	tokenData := &auth.EmailVerificationTokenData{
		Email:    req.Email,
		Token:    token,
		ExpiryAt: expiryAt,
	}
	err = serv.repo.SetEmailVerificationToken(ctx, tokenData)
	if err != nil {
		serv.log.Error("Failed to set email reset token", zap.String("email", req.Email), zap.Error(err))
		return &apperror.APIError{
			Code:    "TOKEN_SET_FAILED",
			Message: "Could not set email reset token. Please try again.",
		}
	}

	// send email verification link
	verificationLink := "/verify-email?token=" + token
	emailData := struct {
		VerificationLink string
		ExpiryInMinutes  int
		Year             string
	}{
		VerificationLink: verificationLink,
		ExpiryInMinutes:  expiryTime,
		Year:             strconv.Itoa(time.Now().Year()),
	}
	err = serv.sender.SendMail(ctx, "./pkg/mailer/update-mail.html", req.Email, "Email Verification", emailData)
	if err != nil {
		serv.log.Error("Failed to send email reset link", zap.String("email", req.Email), zap.Error(err))
		return &apperror.APIError{
			Code:    "SEND_RESET_EMAIL_FAILED",
			Message: "Failed to send email reset link. Please try again.",
		}
	}
	serv.log.Info("Email reset link sent successfully", zap.String("email", req.Email))
	return nil
}

func (serv *AuthService) VerifyEmailChangeRequest(ctx context.Context, req *auth.VerifyEmailRequest) *apperror.APIError {
	tokenData, err := serv.repo.GetEmailVerificationToken(ctx, req.Token)
	if err != nil {
		if err == apperror.ErrTokenInvalid {
			return &apperror.APIError{
				Code:    "INVALID_OR_EXPIRED_TOKEN",
				Message: "Email reset token is invalid or expired.",
			}
		}
		serv.log.Error("Failed to get email reset token", zap.String("token", req.Token), zap.Error(err))
		return &apperror.APIError{
			Code:    "TOKEN_FETCH_FAILED",
			Message: "Something went wrong. Please try again.",
		}
	}

	// get current email from context
	currentEmail := utils.GetEmailFromContext(ctx)
	if currentEmail == "" {
		return &apperror.APIError{
			Code:    "EMAIL_NOT_FOUND",
			Message: "Email not found in context.",
		}
	}

	// update email
	err = serv.repo.UpdateEmail(ctx, currentEmail, tokenData)
	if err != nil {
		serv.log.Error("Failed to update email", zap.String("email", tokenData.Email), zap.Error(err))
		return &apperror.APIError{
			Code:    "EMAIL_UPDATE_FAILED",
			Message: "Failed to update email. Please try again.",
		}
	}

	serv.log.Info("Email updated successfully", zap.String("email", tokenData.Email))
	return nil
}

// refresh token
func (serv *AuthService) VerifyRefreshToken(ctx context.Context, token string) (*auth.RefreshTokenResponse, *apperror.APIError) {

	refreshToken, userData, err := serv.repo.GetRefreshToken(ctx, token)
	if err != nil {
		serv.log.Error("Failed to get refresh token", zap.String("token", token), zap.Error(err))
		return nil, &apperror.APIError{
			Code:    "REFRESH_TOKEN_NOT_FOUND",
			Message: "Refresh token not found.",
		}
	}

	refreshToken.UserID = userData.UserID
	err = utils.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, &apperror.APIError{
			Code:    "REFRESH_TOKEN_INVALID",
			Message: err.Error(),
		}
	}

	accessToken, err := utils.IssueJWT(userData.UserID, userData.Email, userData.Role, userData.IsVerified, serv.log)
	if err != nil {
		return nil, &apperror.APIError{
			Code:    "ACCESS_TOKEN_ISSUANCE_FAILED",
			Message: "Failed to issue access token.",
		}
	}
	newrefreshToken, expiryAt, err := utils.GenerateTokenWithExpiry(32, 15*1440)
	if err != nil {
		return nil, &apperror.APIError{
			Code:    "REFRESH_TOKEN_GENERATION_FAILED",
			Message: "Failed to generate refresh token.",
		}
	}
	err = serv.repo.SetRefreshToken(ctx, &auth.RefreshToken{
		UserID:   userData.UserID,
		Token:    newrefreshToken,
		ExpiryAt: expiryAt,
		Revoked:  false,
	})
	if err != nil {
		serv.log.Error("Failed to set refresh token", zap.String("token", newrefreshToken), zap.Error(err))
		return nil, &apperror.APIError{
			Code:    "REFRESH_TOKEN_SET_FAILED",
			Message: "Failed to set refresh token.",
		}
	}
	return &auth.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newrefreshToken,
	}, nil
}

// revoke refresh token
func (serv *AuthService) RevokeRefreshToken(ctx context.Context, token string) *apperror.APIError {
	err := serv.repo.RevokeRefreshToken(ctx, token)
	if err != nil {
		serv.log.Error("Failed to revoke refresh token", zap.String("token", token), zap.Error(err))
		return &apperror.APIError{
			Code:    "REFRESH_TOKEN_REVOKING_FAILED",
			Message: "Failed to revoke refresh token.",
		}
	}
	return nil
}
