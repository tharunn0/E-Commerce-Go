package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/tharunn0/E-Commerce-Go/internal/domain"
	"github.com/tharunn0/E-Commerce-Go/internal/service"
	"github.com/tharunn0/E-Commerce-Go/internal/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"google.golang.org/api/idtoken"
)

type UserHandler struct {
	service  *service.UserService
	logger   *zap.Logger
	authserv *service.AuthService
	oauth    *oauth2.Config
}

func NewUserHandler(srv *service.UserService, authserv *service.AuthService, log *zap.Logger, oauth *oauth2.Config) *UserHandler {
	return &UserHandler{
		service:  srv,
		logger:   log,
		authserv: authserv,
		oauth:    oauth,
	}
}

func (h *UserHandler) RegisterUser(c *gin.Context) {

	ctx := context.Background()
	var req domain.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid registration payload", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "Please provide valid registration details.",
		})
		return
	}

	h.logger.Debug("registering user", zap.String("service", "UserHandler"), zap.Any("request", req))

	apiErr := h.service.RegisterUser(ctx, &req)
	if apiErr != nil {
		c.JSON(apiErr.Status, gin.H{
			"error":   apiErr.Code,
			"message": apiErr.Message,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User has been successfully created",
	})
}

func (h *UserHandler) LoginUser(c *gin.Context) {

	ctx := c.Request.Context()
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid login payload", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "Invalid request payload.",
		})
		return
	}

	user, apiErr := h.service.LoginUser(ctx, &req)
	if apiErr != nil {
		h.logger.Warn("login failed",
			zap.String("email", req.Email),
			zap.String("code", apiErr.Code),
			zap.String("msg", apiErr.Message),
		)
		c.JSON(apiErr.Status, gin.H{
			"error":   apiErr.Code,
			"message": apiErr.Message,
		})
		return
	}

	h.logger.Info("user logged in successfully",
		zap.Int64("user_id", user.User.ID),
		zap.String("email", user.User.Email),
		zap.String("role", user.User.Role),
	)

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) LogoutUser(c *gin.Context) {
	ctx := c.Request.Context()
	header := c.GetHeader("Authorization")
	token, err := utils.ExtractAuthToken(header)
	if err != nil {
		h.logger.Error("Failed to extract refresh token", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "Please provide a valid refresh token.",
		})
		return
	}

	apiErr := h.authserv.RevokeRefreshToken(ctx, token)
	if apiErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   apiErr.Code,
			"message": apiErr.Message,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User logged out successfully",
	})
}

func (h *UserHandler) SendOTP(c *gin.Context) {

	type SendOTPReq struct {
		Email string `json:"email" validate:"required,email"`
	}
	var req SendOTPReq

	ctx := c.Request.Context()
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid send OTP request payload", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "Please provide a valid email address.",
		})
		return
	}

	apiErr := h.authserv.SendOTP(ctx, req.Email)
	if apiErr != nil {
		h.logger.Warn("OTP send failed",
			zap.String("email", req.Email),
			zap.String("code", apiErr.Code),
			zap.String("message", apiErr.Message),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   apiErr.Code,
			"message": apiErr.Message,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "OTP sent successfully to your registered email",
	})
}

func (h *UserHandler) VerifyOTP(c *gin.Context) {

	var req domain.VerifyOTPReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "Please provide valid email and OTP.",
		})
		return
	}

	apiErr := h.authserv.VerifyOTP(req.Email, req.OTP)
	if apiErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   apiErr.Code,
			"message": apiErr.Message,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "OTP verified successfully",
	})
}

func (h *UserHandler) SendPasswordResetLink(c *gin.Context) {
	ctx := context.Background()
	var req domain.PasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid password reset request payload", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "Please provide a valid email address.",
		})
		return
	}

	apiErr := h.authserv.SendPasswordResetLink(ctx, &req)
	if apiErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   apiErr.Code,
			"message": apiErr.Message,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "If an account with this email exists, a reset link has been sent.",
	})
}

func (h *UserHandler) ResetPassword(c *gin.Context) {
	ctx := context.Background()

	token := c.Query("token")
	if token == "" {
		h.logger.Warn("invalid password reset request payload")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "Please provide a valid token.",
		})
		return
	}
	var req domain.PasswordResetData
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid password reset request payload", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "Please provide a valid email and token.",
		})
		return
	}

	req.Token = token

	apiErr := h.authserv.ResetPassword(ctx, &req)
	if apiErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   apiErr.Code,
			"message": apiErr.Message,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset successfully",
	})
}

func (h *UserHandler) GoogleSignIn(c *gin.Context) {

	state := utils.GenerateState()
	if state == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "GOOGLE_SIGN_IN_FAILED",
			"message": "Could not generate state. Please try again.",
		})
		return
	}
	c.SetCookie("oauth_state", state, 3600, "/", "127.0.0.1", false, true)

	redirectURL := h.oauth.AuthCodeURL(state, oauth2.AccessTypeOffline)

	c.Redirect(http.StatusFound, redirectURL)

}

func (h *UserHandler) GoogleCallback(c *gin.Context) {

	code := c.Query("code")
	state, err := c.Cookie("oauth_state")
	if err != nil {
		h.logger.Error("error getting oauth state", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "GOOGLE_SIGN_IN_FAILED",
			"message": "Could not retrieve necessary state. Please try again.",
		})
		return
	}
	if state == "" {
		c.Redirect(http.StatusFound, "/api/v1/auth/users/google?error=state_mismatch")
		return
	}

	if state != c.Query("state") {
		c.Redirect(http.StatusFound, "/api/v1/auth/users/google?error=state_mismatch")
		return
	}

	token, err := h.oauth.Exchange(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "GOOGLE_TOKEN_EXCHANGE_FAILED",
			"message": "Something went wrong with Google token exchange. Please try again.",
		})
		return
	}

	rawIDToken := token.Extra("id_token").(string)

	payload, err := idtoken.Validate(c.Request.Context(), rawIDToken, h.oauth.ClientID)
	if err != nil {
		h.logger.Error("error validating google token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "GOOGLE_TOKEN_VALIDATION_FAILED",
			"message": "Something went wrong with Google token validation. Please try again.",
		})
		return
	}

	email, _ := payload.Claims["email"].(string)
	verified, _ := payload.Claims["email_verified"].(bool)
	firstName, _ := payload.Claims["given_name"].(string)
	lastName, _ := payload.Claims["family_name"].(string)
	sub, _ := payload.Claims["sub"].(string)

	req := &domain.GoogleSignInRequest{
		Email:     email,
		Token:     rawIDToken,
		Verified:  verified,
		FirstName: firstName,
		LastName:  lastName,
		Sub:       sub,
	}

	user, apiErr := h.service.OAuthSignIn(c.Request.Context(), req)
	if apiErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   apiErr.Code,
			"message": apiErr.Message,
		})
		return
	}

	c.JSON(http.StatusOK, user)

}

func (h *UserHandler) EmailChangeRequest(c *gin.Context) {
	ctx := c.Request.Context()
	var req domain.UpdateEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid update email request payload", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "Please provide a valid email address.",
		})
		return
	}

	apiErr := h.authserv.SendEmailChangeLink(ctx, &req)
	if apiErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   apiErr.Code,
			"message": apiErr.Message,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Email reset link sent successfully",
	})
}

func (h *UserHandler) VerifyEmailChangeRequest(c *gin.Context) {
	ctx := c.Request.Context()
	var req domain.VerifyEmailRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "Url contains invalid token.",
		})
		return
	}

	apiErr := h.authserv.VerifyEmailChangeRequest(ctx, &req)
	if apiErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   apiErr.Code,
			"message": apiErr.Message,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Email updated successfully",
	})
}

func (h *UserHandler) RefreshToken(c *gin.Context) {
	ctx := c.Request.Context()
	var token string

	header := c.GetHeader("Authorization")
	token, err := utils.ExtractAuthToken(header)
	if err != nil {
		h.logger.Error("Failed to extract refresh token", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "Please provide a valid refresh token.",
		})
		return
	}

	authTokens, apiErr := h.authserv.VerifyRefreshToken(ctx, token)
	if apiErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   apiErr.Code,
			"message": apiErr.Message,
		})
		return
	}
	c.JSON(http.StatusOK, authTokens)
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	ctx := c.Request.Context()

	userProfile, apiErr := h.service.GetUserProfile(ctx)
	if apiErr != nil {
		c.JSON(apiErr.Status, gin.H{
			"error":   apiErr.Code,
			"message": apiErr.Message,
		})
		return
	}
	c.JSON(http.StatusOK, userProfile)
}

func (h *UserHandler) UpdateUserProfile(c *gin.Context) {
	ctx := c.Request.Context()
	var req domain.UpdateUserProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid update user profile request payload", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "Please provide a valid user profile.",
		})
		return
	}
	userProfile, apiErr := h.service.UpdateUserProfile(ctx, &req)
	if apiErr != nil {
		c.JSON(apiErr.Status, gin.H{
			"error":   apiErr.Code,
			"message": apiErr.Message,
		})
		return
	}
	h.logger.Info("user profile updated successfully", zap.Int64("user_id", userProfile.ID), zap.String("email", userProfile.Email))
	c.JSON(http.StatusOK, userProfile)
}

//user address

func (h *UserHandler) CreateUserAddress(c *gin.Context) {
	ctx := c.Request.Context()
	var req domain.UserAddress
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid create user address request payload", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "Please provide a valid user address.",
		})
		return
	}
	if err := utils.ValidateUserAddressValues(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": err.Error(),
		})
		return
	}
	apiErr := h.service.CreateUserAddress(ctx, &req)
	if apiErr != nil {
		c.JSON(apiErr.Status, gin.H{
			"error":   apiErr.Code,
			"message": apiErr.Message,
		})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"message": "User address created successfully",
	})
}

func (h *UserHandler) GetUserAddresses(c *gin.Context) {

	ctx := c.Request.Context()

	addresses, defaultAddressID, apiErr := h.service.GetUserAddresses(ctx)
	if apiErr != nil {
		c.JSON(apiErr.Status, gin.H{
			"error":   apiErr.Code,
			"message": apiErr.Message,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"defaultAddressID": defaultAddressID,
		"addresses":        addresses,
	})
}

func (h *UserHandler) UpdateDefaultUserAddress(c *gin.Context) {
	ctx := c.Request.Context()
	addressID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "Address ID not found",
		})
		return
	}
	apiErr := h.service.UpdateDefaultUserAddress(ctx, addressID)
	if apiErr != nil {
		c.JSON(apiErr.Status, gin.H{
			"error":   apiErr.Code,
			"message": apiErr.Message,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"defaultAddressID": addressID,
		"message":          "Default user address updated successfully",
	})
}

func (h *UserHandler) UpdateUserAddress(c *gin.Context) {
	ctx := c.Request.Context()
	addressID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "Address ID not found",
		})
		return
	}

	var req domain.UpdateUserAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid update user address request payload", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "Please provide a valid user address.",
		})
		return
	}

	req.ID = addressID
	updatedAddress, apiErr := h.service.UpdateUserAddress(ctx, &req)
	if apiErr != nil {
		c.JSON(apiErr.Status, gin.H{
			"error":   apiErr.Code,
			"message": apiErr.Message,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"address": updatedAddress,
		"message": "User address updated successfully",
	})
}

func (h *UserHandler) DeleteUserAddress(c *gin.Context) {
	ctx := c.Request.Context()
	addressID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "Address ID not found",
		})
		return
	}
	apiErr := h.service.DeleteUserAddress(ctx, addressID)
	if apiErr != nil {
		c.JSON(apiErr.Status, gin.H{
			"error":   apiErr.Code,
			"message": apiErr.Message,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "User address deleted successfully",
	})
}

// wallets handlers

func (h *UserHandler) GetWallet(c *gin.Context) {
	ctx := c.Request.Context()
	wallet, apiErr := h.service.GetWallet(ctx)
	if apiErr != nil {
		c.JSON(apiErr.Status, gin.H{
			"error":   apiErr.Code,
			"message": apiErr.Message,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"wallet": wallet,
	})
}

func (h *UserHandler) GetWalletTransactions(c *gin.Context) {
	ctx := c.Request.Context()

	var req domain.TransactionFilter
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "Please provide a valid transaction filter.",
		})
		return
	}

	walletTransactions, apiErr := h.service.GetWalletTransactions(ctx, &req)
	if apiErr != nil {
		c.JSON(apiErr.Status, gin.H{
			"error":   apiErr.Code,
			"message": apiErr.Message,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"walletTransactions": walletTransactions,
	})
}
