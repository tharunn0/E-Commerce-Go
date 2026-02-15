package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tharunn0/E-Commerce-Go/internal/domain"
	"github.com/tharunn0/E-Commerce-Go/internal/service"
	"go.uber.org/zap"
)

type CouponHandler struct {
	serv *service.CouponService
	log  *zap.Logger
}

func NewCouponHandler(serv *service.CouponService, log *zap.Logger) *CouponHandler {
	return &CouponHandler{
		serv: serv,
		log:  log,
	}
}

func (h *CouponHandler) CreateCoupon(c *gin.Context) {

	ctx := c.Request.Context()

	var req domain.CreateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	created, apierr := h.serv.CreateCoupon(ctx, &req)
	if apierr != nil {
		c.JSON(apierr.Status, gin.H{"error": apierr.Code, "message": apierr.Message})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Coupon created successfully", "status": "success", "data": created})
}
