package handler

import (
	"net/http"
	"strconv"

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

// CHECKOUT HANDLERS

// checkout cart
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

// checkout product variant
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
	productVariant, err := h.serv.CheckoutProductVariant(ctx, &req)
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

// ORDER HANDLERS

// create order
func (h *OrderHandler) CreateOrder(c *gin.Context) {

	ctx := c.Request.Context()

	var req domain.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "BAD_REQUEST",
			"message": "Invalid request body.",
		})
		return
	}

	// validate and create order

	if req.ProductVariant != nil {
		return
	}

	order, stockerr, err := h.serv.CreateOrderFromCart(ctx, &req)
	if err != nil {
		c.JSON(err.Status, gin.H{
			"error":   err.Code,
			"message": err.Message,
		})
		return
	}
	if stockerr != nil {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "NOT_ENOUGH_STOCK",
			"message": stockerr,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Order created successfully",
		"order":   order,
	})

}

// get orders
func (h *OrderHandler) GetOrders(c *gin.Context) {

	ctx := c.Request.Context()

	orders, err := h.serv.GetOrders(ctx)
	if err != nil {
		c.JSON(err.Status, gin.H{
			"error":   err.Code,
			"message": err.Message,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Orders fetched successfully",
		"orders":  orders,
	})
}

// get order by id
func (h *OrderHandler) GetOrderByID(c *gin.Context) {

	ctx := c.Request.Context()

	orderID := c.Param("order_id")

	order, err := h.serv.GetOrderByID(ctx, orderID)
	if err != nil {
		c.JSON(err.Status, gin.H{
			"error":   err.Code,
			"message": err.Message,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Order fetched successfully",
		"order":   order,
	})
}

func (h *OrderHandler) CancelOrderItem(c *gin.Context) {

	ctx := c.Request.Context()

	orderID := c.Param("order_id")
	variantIDstr := c.Param("variant_id")

	variantID, err := strconv.ParseInt(variantIDstr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "BAD_REQUEST",
			"message": "Invalid variant ID.",
		})
		return
	}

	order, apierr := h.serv.CancelOrderItem(ctx, orderID, variantID)
	if apierr != nil {
		c.JSON(apierr.Status, gin.H{
			"error":   apierr.Code,
			"message": apierr.Message,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Order item cancelled successfully",
		"order":   order,
	})
}

// ORDER ADMIN HANDLERS

// ship order
func (h *OrderHandler) ShipOrder(c *gin.Context) {

	ctx := c.Request.Context()

	orderID := c.Param("id")

	order, apierr := h.serv.ShipOrder(ctx, orderID)
	if apierr != nil {
		c.JSON(apierr.Status, gin.H{
			"error":   apierr.Code,
			"message": apierr.Message,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Order shipped successfully",
		"order":   order,
	})
}

// deliver order
func (h *OrderHandler) DeliverOrder(c *gin.Context) {

	ctx := c.Request.Context()

	orderID := c.Param("id")

	order, apierr := h.serv.DeliverOrder(ctx, orderID)
	if apierr != nil {
		c.JSON(apierr.Status, gin.H{
			"error":   apierr.Code,
			"message": apierr.Message,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Order delivered successfully",
		"order":   order,
	})
}

// list all orders
func (h *OrderHandler) ListAllOrders(c *gin.Context) {

	ctx := c.Request.Context()

	var filter domain.OrderFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "BAD_REQUEST",
			"message": "Invalid request query.",
		})
		return
	}

	orders, apierr := h.serv.ListAllOrders(ctx, &filter)
	if apierr != nil {
		c.JSON(apierr.Status, gin.H{
			"error":   apierr.Code,
			"message": apierr.Message,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Orders fetched successfully",
		"orders":  orders,
	})
}
