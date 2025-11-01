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

type ProductHandler struct {
	service *service.ProductService
	log     *zap.Logger
}

func NewProductHandler(service *service.ProductService, log *zap.Logger) *ProductHandler {
	return &ProductHandler{service: service, log: log}
}

// Brand operations
func (h *ProductHandler) CreateBrand(c *gin.Context) {
	ctx := context.Background()
	var brand domain.CreateBrandRequest
	if err := c.ShouldBindJSON(&brand); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Provide valid brand details"})
		return
	}
	createdBrand, apierr := h.service.CreateBrand(ctx, &brand)

	if apierr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierr.Code, "message": apierr.Message})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "success", "brand": createdBrand})
}

func (h *ProductHandler) GetBrands(c *gin.Context) {
	ctx := c.Request.Context()
	brands, apierr := h.service.GetAllBrands(ctx)
	if apierr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierr.Code, "message": apierr.Message})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "brands": brands})
}

func (h *ProductHandler) GetBrandByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid brand ID"})
		return
	}
	brand, apierr := h.service.GetBrandByID(ctx, idInt)
	if apierr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierr.Code, "message": apierr.Message})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "brand": brand})
}

func (h *ProductHandler) UpdateBrand(c *gin.Context) {
	ctx := context.Background()

	var updateBrandRequest domain.UpdateBrandRequest
	if err := c.ShouldBindJSON(&updateBrandRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedBrand, apierr := h.service.UpdateBrand(ctx, &updateBrandRequest)
	if apierr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierr.Code, "message": apierr.Message})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "brand": updatedBrand})
}

func (h *ProductHandler) DeleteBrand(c *gin.Context) {
	ctx := context.Background()
	id := c.Param("id")
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid brand ID"})
		return
	}
	apierr := h.service.DeleteBrand(ctx, idInt)
	if apierr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierr.Code, "message": apierr.Message})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Brand deleted successfully"})
}
