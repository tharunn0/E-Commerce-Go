package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tharunn0/E-Commerce-Go/internal/config"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/auth"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/auth/mocks"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestAuthService_SendOTP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockAuthRepository(ctrl)
	mockSender := mocks.NewMockEmailSender(ctrl)
	logger := zap.NewNop()
	cfg := &config.SecuritySettings{
		OTPExpiryMinutes: 5,
	}

	authService := NewAuthService(mockRepo, mockSender, logger, cfg)

	t.Run("successfully sends OTP", func(t *testing.T) {
		ctx := context.Background()
		email := "user@example.com"

		mockRepo.EXPECT().InsertOTP(ctx, gomock.Any()).Return(nil)
		mockSender.EXPECT().SendMail(ctx, gomock.Any(), email, "Email Verification", gomock.Any()).Return(nil)

		apiErr := authService.SendOTP(ctx, email)
		assert.Nil(t, apiErr)
	})

	t.Run("fails when email sender returns error", func(t *testing.T) {
		ctx := context.Background()
		email := "user@example.com"

		mockRepo.EXPECT().InsertOTP(ctx, gomock.Any()).Return(nil)
		mockSender.EXPECT().SendMail(ctx, gomock.Any(), email, "Email Verification", gomock.Any()).Return(errors.New("SMTP error"))

		apiErr := authService.SendOTP(ctx, email)
		assert.NotNil(t, apiErr)
		assert.Equal(t, "OTP_EMAIL_FAILED", apiErr.Code)
	})
}

func TestAuthService_SendPasswordResetLink(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockAuthRepository(ctrl)
	mockSender := mocks.NewMockEmailSender(ctrl)
	logger := zap.NewNop()
	cfg := &config.SecuritySettings{
		PasswordResetExpiry: 15,
	}

	authService := NewAuthService(mockRepo, mockSender, logger, cfg)

	t.Run("successfully sends reset link", func(t *testing.T) {
		ctx := context.Background()
		req := &auth.PasswordResetRequest{Email: "user@example.com"}

		mockRepo.EXPECT().SetPasswordResetToken(ctx, gomock.Any()).Return(nil)
		mockSender.EXPECT().SendMail(ctx, gomock.Any(), req.Email, "Password Reset Request", gomock.Any()).Return(nil)

		apiErr := authService.SendPasswordResetLink(ctx, req)
		assert.Nil(t, apiErr)
	})

	t.Run("invalid email format", func(t *testing.T) {
		ctx := context.Background()
		req := &auth.PasswordResetRequest{Email: "invalid-email"}

		mockSender.EXPECT().SendMail(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

		apiErr := authService.SendPasswordResetLink(ctx, req)
		assert.NotNil(t, apiErr)
		assert.Equal(t, "INVALID_EMAIL", apiErr.Code)
	})
}
