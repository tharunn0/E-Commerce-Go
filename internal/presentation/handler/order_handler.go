package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tharunn0/E-Commerce-Go/internal/domain"
	"github.com/tharunn0/E-Commerce-Go/internal/service"
	"go.uber.org/zap"
)

type OrderHandler struct {
	serv *service.OrderService

	log *zap.Logger
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

// get user orders
func (h *OrderHandler) GetUserOrders(c *gin.Context) {

	ctx := c.Request.Context()

	orders, err := h.serv.GetUserOrders(ctx)
	if err != nil {
		c.JSON(err.Status, gin.H{
			"error":   err.Code,
			"message": err.Message,
		})
		return
	}

	if orders == nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "No orders found for user",
			"orders":  []int{},
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

// cancel order item
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

// cancel order
func (h *OrderHandler) CancelOrder(c *gin.Context) {

	ctx := c.Request.Context()

	orderID := c.Param("order_id")

	var req domain.CancelOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "BAD_REQUEST",
			"message": "Invalid request body.",
		})
		return
	}

	req.OrderID = orderID

	order, apierr := h.serv.CancelOrder(ctx, &req)
	if apierr != nil {
		c.JSON(apierr.Status, gin.H{
			"error":   apierr.Code,
			"message": apierr.Message,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Order cancelled successfully",
		"order":   order,
	})
}

func (h *OrderHandler) ReturnOrderItemRequest(c *gin.Context) {

	ctx := c.Request.Context()

	fmt.Println("return item endpoint called")

	var req domain.ReturnOrderItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "BAD_REQUEST",
			"message": "Invalid request body.",
		})
		return
	}

	orderId := c.Param("order_id")
	variantId := c.Param("variant_id")

	variantIdInt, err := strconv.ParseInt(variantId, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "BAD_REQUEST",
			"message": "Invalid variant ID.",
		})
		return
	}

	req.OrderID = orderId
	req.ItemID = variantIdInt

	apierr := h.serv.ReturnOrderItemRequest(ctx, &req)
	if apierr != nil {
		c.JSON(apierr.Status, gin.H{
			"error":   apierr.Code,
			"message": apierr.Message,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Return order item request created successfully",
	})
}

func (h *OrderHandler) ReturnOrderRequest(c *gin.Context) {

	ctx := c.Request.Context()

	var req domain.ReturnOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "BAD_REQUEST",
			"message": "Invalid request body.",
		})
		return
	}

	orderID := c.Param("order_id")
	req.OrderID = orderID

	if req.OrderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "BAD_REQUEST",
			"message": "Invalid order ID.",
		})
		return
	}

	apierr := h.serv.ReturnOrderRequest(ctx, &req)
	if apierr != nil {
		c.JSON(apierr.Status, gin.H{
			"error":   apierr.Code,
			"message": apierr.Message,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Order return request created successfully",
	})
}

// ORDER ADMIN HANDLERS

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

// update order status
func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {

	ctx := c.Request.Context()

	var req domain.OrderStatusUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "BAD_REQUEST",
			"message": "Invalid request body.",
		})
		return
	}

	order := c.Param("order_id")
	req.OrderID = order

	_, apierr := h.serv.UpdateOrderStatus(ctx, req.OrderID, domain.ShipmentStatus(req.Status))
	if apierr != nil {
		c.JSON(apierr.Status, gin.H{
			"error":   apierr.Code,
			"message": apierr.Message,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Order status updated successfully",
		"order":   order,
	})
}

// order returns

func (h *OrderHandler) ListAllReturns(c *gin.Context) {
	ctx := c.Request.Context()

	var filter domain.ReturnFilter

	st := c.Query("status")
	fmt.Println("status", st)

	if err := c.ShouldBindQuery(&filter); err != nil {
		h.log.Warn("failed to bind query", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "BAD_REQUEST",
			"message": "Invalid request query.",
		})
		return
	}

	returns, apierr := h.serv.ListReturnRequests(ctx, &filter)
	if apierr != nil {
		c.JSON(apierr.Status, gin.H{
			"error":   apierr.Code,
			"message": apierr.Message,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Returns fetched successfully",
		"returns": returns,
	})
}

func (h *OrderHandler) GetReturnRequest(c *gin.Context) {
	ctx := c.Request.Context()

	returnID := c.Param("return_id")

	returnRequest, apierr := h.serv.GetReturnRequest(ctx, returnID)
	if apierr != nil {
		c.JSON(apierr.Status, gin.H{
			"error":   apierr.Code,
			"message": apierr.Message,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Return request fetched successfully",
		"return":  returnRequest,
	})
}

func (h *OrderHandler) UpdateReturnRequestStatus(c *gin.Context) {

	ctx := c.Request.Context()

	returnID := c.Param("return_id")

	returnRequest, apierr := h.serv.GetReturnRequest(ctx, returnID)
	if apierr != nil {
		c.JSON(apierr.Status, gin.H{
			"error":   apierr.Code,
			"message": apierr.Message,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Return request fetched successfully",
		"return":  returnRequest,
	})
}

func (h *OrderHandler) GetReturnByID(c *gin.Context) {

}
