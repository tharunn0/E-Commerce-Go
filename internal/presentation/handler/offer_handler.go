package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/promotion"
	"github.com/tharunn0/E-Commerce-Go/internal/service"
	"go.uber.org/zap"
)

type OfferHandler struct {
	serv *service.OfferService
	log  *zap.Logger
}

func NewOfferHandler(serv *service.OfferService, log *zap.Logger) *OfferHandler {
	return &OfferHandler{serv: serv, log: log}
}

func (h *OfferHandler) CreateOffer(c *gin.Context) {

	ctx := c.Request.Context()

	var req promotion.CreateOfferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("Failed to bind JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if h.serv == nil {
		h.log.Error("Failed to initialize offer service", zap.String("function", "CreateOffer"))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if apierr := h.serv.CreateOffer(ctx, &req); apierr != nil {
		c.JSON(apierr.Status, gin.H{"code": apierr.Code, "error": apierr.Message})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Offer created successfully"})

}

func (h *OfferHandler) GetAllProductOffers(c *gin.Context) {

	ctx := c.Request.Context()

	offers, apierr := h.serv.GetAllProductOffers(ctx)
	if apierr != nil {
		c.JSON(apierr.Status, gin.H{"code": apierr.Code, "error": apierr.Message})
		return
	}
	c.JSON(http.StatusOK, gin.H{"offers": offers})
}

func (h *OfferHandler) GetAllCategoryOffers(c *gin.Context) {

	ctx := c.Request.Context()

	offers, apierr := h.serv.GetAllCategoryOffers(ctx)
	if apierr != nil {
		c.JSON(apierr.Status, gin.H{"code": apierr.Code, "error": apierr.Message})
		return
	}
	c.JSON(http.StatusOK, gin.H{"offers": offers})
}
