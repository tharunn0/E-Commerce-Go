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

// product handler
// Product operations
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	ctx := context.Background()
	var createProductRequest domain.CreateProductRequest
	if err := c.ShouldBindJSON(&createProductRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Provide valid product details"})
		return
	}
	h.log.Debug("creating product", zap.Any("request", createProductRequest))
	createdProduct, apierr := h.service.CreateProduct(ctx, &createProductRequest)
	if apierr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierr.Code, "message": apierr.Message})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"product": createdProduct})
}

func (h *ProductHandler) GetProducts(c *gin.Context) {
	ctx := c.Request.Context()
	var filter domain.ProductFilter

	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "INVALID_QUERY_PARAMETERS", "message": err})
	}

	// filter.Page, _ = strconv.Atoi(c.Query("page"))
	// filter.Limit, _ = strconv.Atoi(c.Query("limit"))
	// filter.Search = c.Query("search")
	// filter.Sort = c.Query("sort")
	// filter.Order = c.Query("order")
	// tmax, _ := strconv.Atoi(c.Query("max_price"))
	// tmin, _ := strconv.Atoi(c.Query("min_price"))
	// filter.MaxPrice, filter.MinPrice = float64(tmax), float64(tmin)

	products, total, apierr := h.service.GetProducts(ctx, &filter)
	if apierr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierr.Code, "message": apierr.Message})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"page":     filter.Page,
		"limit":    filter.Limit,
		"total":    total,
		"products": products,
	})
}

func (h *ProductHandler) GetProductByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}
	product, apierr := h.service.GetProductByID(ctx, idInt)
	if apierr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierr.Code, "message": apierr.Message})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "product": product})
}

func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	ctx := c.Request.Context()
	var updateProductRequest domain.UpdateProductRequest
	if err := c.ShouldBindJSON(&updateProductRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updatedProduct, apierr := h.service.UpdateProduct(ctx, &updateProductRequest)
	if apierr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierr.Code, "message": apierr.Message})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "product": updatedProduct})
}

func (h *ProductHandler) UpdateProductStatus(c *gin.Context) {
	ctx := c.Request.Context()
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req domain.ProductStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invaild product status"})
		return
	}
	req.ID = id

	if apierr := h.service.UpdateProductStatus(ctx, &req); apierr != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": apierr.Code, "message": apierr.Message})
		return
	}

	c.JSON(http.StatusNoContent, gin.H{"message": "Product status updated !"})
}

func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	ctx := context.Background()
	id := c.Param("id")
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}
	apierr := h.service.DeleteProduct(ctx, idInt)
	if apierr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierr.Code, "message": apierr.Message})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Product deleted successfully"})
}

// Product variant operations
func (h *ProductHandler) CreateProductVariant(c *gin.Context) {
	ctx := context.Background()
	var createProductVariantRequest domain.CreateProductVariantRequest
	if err := c.ShouldBindJSON(&createProductVariantRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Provide valid product variant details"})
		return
	}

	createdProductVariant, apierr := h.service.CreateProductVariant(ctx, &createProductVariantRequest)
	if apierr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierr.Code, "message": apierr.Message})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"product_variant": createdProductVariant})
}

func (h *ProductHandler) GetProductVariantByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product variant ID"})
		return
	}
	productVariant, apierr := h.service.GetProductVariantByID(ctx, idInt)
	if apierr != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": apierr.Code, "message": apierr.Message})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "product_variant": productVariant})
}

func (h *ProductHandler) GetVariantsByProductID(c *gin.Context) {
	ctx := c.Request.Context()
	productID := c.Param("id")
	productIDInt, err := strconv.ParseInt(productID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	variants, apierr := h.service.GetVariantsByProductID(ctx, productIDInt)
	if apierr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierr.Code, "message": apierr.Message})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "variants": variants})
}

func (h *ProductHandler) UpdateProductVariant(c *gin.Context) {
	ctx := c.Request.Context()

	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var updateProductVariantRequest domain.UpdateProductVariantRequest
	updateProductVariantRequest.ID = id
	if err := c.ShouldBindJSON(&updateProductVariantRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updatedProductVariant, apierr := h.service.UpdateProductVariant(ctx, &updateProductVariantRequest)
	if apierr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierr.Code, "message": apierr.Message})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "product_variant": updatedProductVariant})
}

func (h *ProductHandler) UpdateVariantStatus(c *gin.Context) {
	ctx := c.Request.Context()
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req domain.VariantStatusRequest
	req.ID = id

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product variant status"})
		return
	}

	err := h.service.UpdateProductVariantStatus(ctx, &req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Code, "message": err.Message})
		return
	}

	c.JSON(http.StatusNoContent, gin.H{"message": "Variant status updated !s"})
}

func (h *ProductHandler) DeleteProductVariant(c *gin.Context) {
	ctx := context.Background()
	id := c.Param("id")
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product variant ID"})
		return
	}

	apierr := h.service.DeleteProductVariant(ctx, idInt)
	if apierr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierr.Code, "message": apierr.Message})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Product variant deleted successfully"})
}

// Attribute operations
func (h *ProductHandler) CreateAttribute(c *gin.Context) {
	ctx := context.Background()
	var createProductVariantRequest domain.CreateAttributeRequest
	if err := c.ShouldBindJSON(&createProductVariantRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Provide valid product variant details"})
		return
	}

	createdAttribute, apierr := h.service.CreateAttribute(ctx, &createProductVariantRequest)
	if apierr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierr.Code, "message": apierr.Message})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"attribute": createdAttribute})
}

func (h *ProductHandler) GetAttributes(c *gin.Context) {
	ctx := c.Request.Context()
	attributes, apierr := h.service.GetAttributes(ctx)
	if apierr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierr.Code, "message": apierr.Message})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "attributes": attributes})
}

func (h *ProductHandler) GetAttributeByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid attribute ID"})
		return
	}
	attribute, apierr := h.service.GetAttributeByID(ctx, idInt)
	if apierr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": apierr.Code, "message": apierr.Message})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "attribute": attribute})
}

func (h *ProductHandler) DeleteAttribute(c *gin.Context) {

	ctx := c.Request.Context()
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var deletereq domain.DeleteAttributeRequest
	deletereq.ID = id

	apierr := h.service.DeleteAttribute(ctx, &deletereq)
	if apierr != nil {

	}
}

func (h *ProductHandler) DeleteAttributeValues(c *gin.Context) {}
