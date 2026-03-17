package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/review"
	"github.com/tharunn0/E-Commerce-Go/internal/service"
	"go.uber.org/zap"
)

type ReviewHandler struct {
	serv *service.ReviewService
	log  *zap.Logger
}

func NewReviewHandler(serv *service.ReviewService, log *zap.Logger) *ReviewHandler {
	return &ReviewHandler{serv: serv, log: log}
}

func (h *ReviewHandler) CreateReview(c *gin.Context) {

	ctx := c.Request.Context()

	var req review.CreateReviewRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": err.Error()})
		return
	}

	productIdstr := c.Param("product_id")
	if productIdstr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Product ID is required"})
		return
	}
	productId, err := strconv.ParseInt(productIdstr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Invalid product ID"})
		return
	}
	req.ProductID = productId

	apierr := h.serv.CreateReview(ctx, &req)
	if apierr != nil {
		c.JSON(apierr.Status, gin.H{"status": "failed", "message": apierr.Message})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Review created successfully"})

}

func (h *ReviewHandler) DeleteReview(c *gin.Context) {

}

func (h *ReviewHandler) GetProductReviews(c *gin.Context) {

	ctx := c.Request.Context()

	var filter review.ReviewFilter

	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": err.Error()})
		return
	}

	productIdstr := c.Param("product_id")
	if productIdstr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Product ID is required"})
		return
	}
	productId, err := strconv.ParseInt(productIdstr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Invalid product ID"})
		return
	}

	productReviews, apierr := h.serv.GetProductReviews(ctx, productId, &filter)
	if apierr != nil {
		c.JSON(apierr.Status, gin.H{"status": "failed", "message": apierr.Message})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": productReviews})

}
