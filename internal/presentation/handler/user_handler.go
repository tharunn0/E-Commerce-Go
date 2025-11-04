package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tharunn0/E-Commerce-Go/internal/domain"
	"github.com/tharunn0/E-Commerce-Go/internal/service"
	"github.com/tharunn0/E-Commerce-Go/internal/utils"

	"go.uber.org/zap"
)

type UserHandler struct {
	service  *service.UserService
	logger   *zap.Logger
	authserv *service.AuthService
}

func NewUserHandler(srv *service.UserService, log *zap.Logger, authserv *service.AuthService) *UserHandler {
	return &UserHandler{
		service:  srv,
		logger:   log,
		authserv: authserv,
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
		c.JSON(http.StatusBadRequest, gin.H{
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

	ctx := context.Background()
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
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   apiErr.Code,
			"message": apiErr.Message,
		})
		return
	}

	// Issue JWT
	token, err := utils.IssueJWT(user.ID, user.Email, user.Role, user.IsVerified, h.logger)
	if len(token) == 0 {
		h.logger.Error("failed to issue jwt", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "TOKEN_GENERATION_FAILED",
			"message": "Could not generate token.",
		})
		return
	}

	resp := domain.LoginResponse{
		Token: token,
	}
	resp.User.ID = user.ID
	resp.User.Email = user.Email
	resp.User.FirstName = user.FirstName
	resp.User.LastName = user.LastName
	resp.User.Role = string(user.Role)

	h.logger.Info("user logged in successfully",
		zap.Int64("user_id", user.ID),
		zap.String("email", user.Email),
		zap.String("role", string(user.Role)),
	)

	c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) SendOTP(c *gin.Context) {

	email := c.Value(domain.KeyEmail).(string)

	apiErr := h.authserv.SendOTP(email)
	if apiErr != nil {
		h.logger.Warn("OTP send failed",
			zap.String("email", email),
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

// func (h *UserHandler) GoogleSignIn(c *gin.Context) {
// 	ctx := context.Background()
// 	var googlereq domain.GoogleSignInRequest
// 	if err := c.ShouldBindJSON(&googlereq); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error":   "INVALID_REQUEST",
// 			"message": "Please provide a valid token.",
// 		})
// 	}

// }

func (h *UserHandler) GetProfile(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.Value(domain.KeyUserID).(int64)

	userProfile, apiErr := h.service.GetUserProfile(ctx, userID)
	if apiErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   apiErr.Code,
			"message": apiErr.Message,
		})
		return
	}
	c.JSON(http.StatusOK, userProfile)
}
