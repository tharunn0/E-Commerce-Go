package handler

import (
	"context"
	"net/http"
	"strconv"

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
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "Invalid request payload.",
		})
		return
	}

	admin, apiErr := h.service.LoginAdmin(&req)
	if apiErr != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   apiErr.Code,
			"message": apiErr.Message,
		})
		return
	}

	c.JSON(http.StatusOK, admin)
}

func (h *AdminHandler) GetAllUsers(c *gin.Context) {
	ctx := context.Background()

	var filters domain.UserFilter
	filters.Status = c.Query("status")
	filters.Role = c.Query("role")
	filters.Search = c.Query("search")
	filters.Page, _ = strconv.Atoi(c.Query("page"))
	filters.Limit, _ = strconv.Atoi(c.Query("limit"))
	users, apierr := h.service.GetAllUsers(ctx, &filters)
	if apierr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierr.Code, "message": apierr.Message})
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": users})
}

func (h *AdminHandler) UpdateUserStatus(c *gin.Context) {
	ctx := context.Background()

	var req domain.UserStatusUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "INVALID_REQUEST", "message": "Invalid request payload."})
		return
	}
	apierr := h.service.UpdateUserStatus(ctx, &req)
	if apierr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierr.Code, "message": apierr.Message})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User status updated successfully"})
}
