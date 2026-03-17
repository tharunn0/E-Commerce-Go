package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/promotion"
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

	var req promotion.CreateCouponRequest
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

func (h *CouponHandler) ListAllCoupons(c *gin.Context) {

	log.Println("list all coupons hit")

	ctx := c.Request.Context()

	var filter promotion.ListCouponsFilter
	_ = c.ShouldBindQuery(&filter)

	log.Println("filter :", filter)

	coupons, apierr := h.serv.ListAllCoupons(ctx, &filter)
	if apierr != nil {
		h.log.Error("coupon listing failed", zap.String("error", apierr.Code), zap.String("message", apierr.Message))
		c.JSON(apierr.Status, gin.H{"error": apierr.Code, "message": apierr.Message})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Coupons fetched successfully", "status": "success", "data": coupons})

}

func (h *CouponHandler) ApplyCoupon(c *gin.Context) {

	log.Println("[handler] apply coupon hit")

	ctx := c.Request.Context()

	var req promotion.ApplyCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	coupon, _, apierr := h.serv.ApplyCoupon(ctx, &req)
	if apierr != nil {
		c.JSON(apierr.Status, gin.H{"error": apierr.Code, "message": apierr.Message})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Coupon applied successfully", "status": "success", "data": coupon})

}
