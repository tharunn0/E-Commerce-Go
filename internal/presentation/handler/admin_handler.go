package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tharunn0/E-Commerce-Go/internal/domain"
	"github.com/tharunn0/E-Commerce-Go/internal/service"
	"go.uber.org/zap"
)

type AdminHandler struct {
	service *service.AdminService
	log     *zap.Logger
}

func NewAdminHandler(srv *service.AdminService, log *zap.Logger) *AdminHandler {
	return &AdminHandler{
		service: srv,
		log:     log,
	}
}

func (h *AdminHandler) LoginUser(c *gin.Context) {
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn("invalid login payload", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "Invalid request payload.",
		})
		return
	}

	admin, apiErr := h.service.LoginAdmin(&req)
	if apiErr != nil {
		h.log.Warn("login failed",
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

	h.log.Info("user logged in successfully",
		zap.Int64("user_id", admin.User.ID),
		zap.String("email", admin.User.Email),
		zap.String("role", string(admin.User.Role)),
	)

	c.JSON(http.StatusOK, admin)
}
