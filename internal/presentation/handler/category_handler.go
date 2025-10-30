package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tharunn0/E-Commerce-Go/internal/domain"
	"github.com/tharunn0/E-Commerce-Go/internal/service"
	"go.uber.org/zap"
)

type CategoryHandler struct {
	serv *service.CategoryService
	log  *zap.Logger
}

func NewCategoryHandler(categoryService *service.CategoryService, log *zap.Logger) *CategoryHandler {
	return &CategoryHandler{serv: categoryService, log: log}
}

func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	ctx := c.Request.Context()
	var category domain.Category
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	createdCategory, apierr := h.serv.CreateCategory(ctx, &category)
	if apierr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierr.Code, "message": apierr.Message})
	}
	c.JSON(http.StatusCreated, gin.H{"category": createdCategory})
}

func (h *CategoryHandler) GetAllCategories(c *gin.Context) {
	ctx := c.Request.Context()

	categories, apierr := h.serv.GetAllCategories(ctx)
	if apierr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierr.Code, "message": apierr.Message})
		return
	}
	c.JSON(http.StatusOK, gin.H{"categories": categories})
}

func (h *CategoryHandler) GetCategoryByID(c *gin.Context) {
	ctx := c.Request.Context()
	c.Request.Context()
	id := c.Param("id")
	intID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
		return
	}
	category, apierr := h.serv.GetCategoryByID(ctx, int64(intID))
	if apierr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierr.Code, "message": apierr.Message})
		return
	}
	c.JSON(http.StatusOK, gin.H{"category": category})
}

func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	intID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
		return
	}
	apierr := h.serv.DeleteCategory(ctx, int64(intID))
	if apierr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierr.Code, "message": apierr.Message})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Category deleted successfully"})
}

func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	ctx := c.Request.Context()

	var updateCategoryRequest domain.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&updateCategoryRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updatedCategory, apierr := h.serv.UpdateCategory(ctx, &updateCategoryRequest)
	if apierr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierr.Code, "message": apierr.Message})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "category": updatedCategory})
}

func (h *CategoryHandler) ToggleCategoryStatus(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	intID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
		return
	}

	var statusRequest domain.CategoryStatusRequest
	if err := c.ShouldBindJSON(&statusRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updatedCategory, apierr := h.serv.ChangeCategoryStatus(ctx, int64(intID), &statusRequest)
	if apierr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierr.Code, "message": apierr.Message})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "category": updatedCategory})
}
