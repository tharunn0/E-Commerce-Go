package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/wishlist"
	"github.com/tharunn0/E-Commerce-Go/internal/service"
	"go.uber.org/zap"
)

type WishlistHandler struct {
	serv *service.WishlistService
	log  *zap.Logger
}

func NewWishlistHandler(wishlistService *service.WishlistService, logger *zap.Logger) *WishlistHandler {
	return &WishlistHandler{
		serv: wishlistService,
		log:  logger,
	}
}

func (h *WishlistHandler) AddToWishlist(c *gin.Context) {
	ctx := c.Request.Context()

	var req wishlist.AddToWishlistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	wish, err := h.serv.AddToWishlist(ctx, req.ProductID)
	if err != nil {
		c.JSON(err.Status, gin.H{"error": err.Code, "message": err.Message})
		return
	}

	c.JSON(200, gin.H{"message": "Added to wishlist", "wishlist": wish})
}

func (h *WishlistHandler) GetWishlist(c *gin.Context) {
	ctx := c.Request.Context()

	wish, err := h.serv.GetWishlistByUserID(ctx)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Code, "message": err.Message})
		return
	}

	c.JSON(200, gin.H{"wishlists": wish})
}

func (h *WishlistHandler) RemoveFromWishlist(c *gin.Context) {
	ctx := c.Request.Context()

	var req wishlist.RemoveFromWishlistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	err := h.serv.RemoveFromWishlist(ctx, req.ProductIDs)
	if err != nil {
		c.JSON(err.Status, gin.H{"error": err.Code, "message": err.Message})
		return
	}
	c.JSON(200, gin.H{"message": "Removed from wishlist"})
}
