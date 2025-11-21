package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tharunn0/E-Commerce-Go/internal/domain"
	"github.com/tharunn0/E-Commerce-Go/internal/service"
	"go.uber.org/zap"
)

type CartHandler struct {
	service *service.CartService
	log     *zap.Logger
}

func NewCartHandler(service *service.CartService, log *zap.Logger) *CartHandler {
	return &CartHandler{service: service, log: log}
}

func (h *CartHandler) AddToCart(c *gin.Context) {
	ctx := c.Request.Context()
	var req domain.AddToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("invalid add to cart request payload", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "Please provide a valid add to cart request.",
		})
		return
	}

	cart, apierr := h.service.AddToCart(ctx, &req)
	if apierr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   apierr.Code,
			"message": apierr.Message,
		})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"message": "Item added to cart successfully",
		"cart":    cart,
	})
}

func (h *CartHandler) GetCart(c *gin.Context) {
	ctx := c.Request.Context()
	cart, apierr := h.service.GetCart(ctx)
	if apierr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   apierr.Code,
			"message": apierr.Message,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"cart": cart,
	})
}

func (h *CartHandler) UpdateCartItemQuantity(c *gin.Context) {
	ctx := c.Request.Context()
	var req domain.UpdateCartItemQuantityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "Please provide a valid update cart item quantity request.",
		})
		return
	}
	updatedCart, apierr := h.service.UpdateCartItemQuantity(ctx, &req)
	if apierr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   apierr.Code,
			"message": apierr.Message,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Cart item quantity updated successfully",
		"cart":    updatedCart,
	})
}

func (h *CartHandler) RemoveCartItem(c *gin.Context) {
	ctx := c.Request.Context()

	productVariantID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "Please provide a valid product variant ID.",
		})
		return
	}
	updatedCart, apierr := h.service.RemoveCartItem(ctx, &domain.RemoveCartItemRequest{
		ProductVariantID: productVariantID,
	})
	if apierr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   apierr.Code,
			"message": apierr.Message,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Cart item removed successfully",
		"cart":    updatedCart,
	})
}

func (h *CartHandler) EmptyCart(c *gin.Context) {
	ctx := c.Request.Context()
	apierr := h.service.EmptyCart(ctx)
	if apierr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   apierr.Code,
			"message": apierr.Message,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Cart emptied successfully",
	})
}
