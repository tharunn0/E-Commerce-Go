package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tharunn0/E-Commerce-Go/internal/domain"
	"github.com/tharunn0/E-Commerce-Go/internal/service"
	"github.com/tharunn0/E-Commerce-Go/internal/utils"
	"go.uber.org/zap"
)

type AdminHandler struct {
	service  *service.AdminService
	log      *zap.Logger
	authserv *service.AuthService
}

func NewAdminHandler(srv *service.AdminService, log *zap.Logger, authserv *service.AuthService) *AdminHandler {
	return &AdminHandler{
		service:  srv,
		log:      log,
		authserv: authserv,
	}
}

func (h *AdminHandler) RegisterAdmin(c *gin.Context) {

	ctx := c.Request.Context()
	var req domain.AdminRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "Invalid request payload.",
		})
	}

	apiErr := h.service.RegisterAdmin(ctx, &req)
	if apiErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   apiErr.Code,
			"message": apiErr.Message,
		})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Admin has been successfully created"})
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

func (h *AdminHandler) RefreshToken(c *gin.Context) {
	ctx := c.Request.Context()
	var token string

	header := c.GetHeader("Authorization")
	token, err := utils.ExtractAuthToken(header)

	if err != nil {
		h.log.Error("Failed to extract refresh token", zap.Error(err))
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

func (h *AdminHandler) GetAllUsers(c *gin.Context) {
	ctx := c.Request.Context()

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
	c.JSON(http.StatusOK, gin.H{
		"users":       users,
		"page":        filters.Page,
		"limit":       filters.Limit,
		"total_users": filters.Total,
	})
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

func (h *AdminHandler) GetUserByID(c *gin.Context) {

	ctx := c.Request.Context()

	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	user, err := h.service.GetUserByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Code, "message": err.Message})
		return
	}

	c.JSON(http.StatusOK, user)

}

func (h *AdminHandler) DeleteUser(c *gin.Context) {

	ctx := c.Request.Context()

	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	err := h.service.DeleteUser(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Code, "message": err.Message})
		return
	}

	c.JSON(http.StatusNoContent, gin.H{"msg": "User successfully deleted"})

}
