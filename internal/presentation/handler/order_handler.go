package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tharunn0/E-Commerce-Go/internal/domain"
	"github.com/tharunn0/E-Commerce-Go/internal/service"
	"go.uber.org/zap"
)

type OrderHandler struct {
	serv *service.OrderService
	log  *zap.Logger
}

func NewOrderHandler(srv *service.OrderService, log *zap.Logger) *OrderHandler {
	return &OrderHandler{serv: srv, log: log}
}

func (h *OrderHandler) CheckoutCart(c *gin.Context) {

	ctx := c.Request.Context()

	var cartReq domain.CartCheckoutRequest
	if err := c.ShouldBindJSON(&cartReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "BAD_REQUEST",
			"message": "Invalid request body.",
		})
		return
	}

	// validate and get cart
	cart, stockErr, err := h.serv.CheckoutCart(ctx, cartReq)
	if err != nil {
		c.JSON(err.Status, gin.H{
			"error":   err.Code,
			"message": err.Message,
		})
		return
	}
	if stockErr != nil {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "NOT_ENOUGH_STOCK",
			"message": stockErr,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cart checked out successfully",
		"cart":    cart,
	})

}

func (h *OrderHandler) CheckoutProductVariant(c *gin.Context) {

	ctx := c.Request.Context()

	// extract request body
	var req domain.ProductVariantCheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "BAD_REQUEST",
			"message": "Invalid request body.",
		})
		return
	}

	// validate and get product variant
	productVariant, err := h.serv.CheckoutProductVariant(ctx, req.ProductVariantID, req.Quantity)
	if err != nil {
		c.JSON(err.Status, gin.H{
			"error":   err.Code,
			"message": err.Message,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Product variant checked out successfully",
		"variant": productVariant,
	})
}
