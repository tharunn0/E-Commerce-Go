package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/order"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/shipping"
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

	var cartReq order.CartCheckoutRequest
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

// ORDER HANDLERS

// create order
func (h *OrderHandler) CreateOrder(c *gin.Context) {

	ctx := c.Request.Context()

	var req order.CreateOrderRequest
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

	orderRes, stockerr, err := h.serv.CreateOrderFromCart(ctx, &req)
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
		"order":   orderRes,
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

	orderRes, err := h.serv.GetOrderByID(ctx, orderID)
	if err != nil {
		c.JSON(err.Status, gin.H{
			"error":   err.Code,
			"message": err.Message,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Order fetched successfully",
		"order":   orderRes,
	})
}

// cancel order
func (h *OrderHandler) CancelOrder(c *gin.Context) {

	ctx := c.Request.Context()

	orderID := c.Param("order_id")

	var req order.CancelOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "BAD_REQUEST",
			"message": "Invalid request body.",
		})
		return
	}

	req.OrderID = orderID

	orderRes, apierr := h.serv.CancelOrder(ctx, &req)
	if apierr != nil {
		c.JSON(apierr.Status, gin.H{
			"error":   apierr.Code,
			"message": apierr.Message,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Order cancelled successfully",
		"order":   orderRes,
	})
}

func (h *OrderHandler) ReturnOrderItemRequest(c *gin.Context) {

	ctx := c.Request.Context()

	fmt.Println("return item endpoint called")

	var req order.ReturnOrderItemRequest
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

	var req order.ReturnOrderRequest
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

	var filter order.OrderFilter
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

	var req order.OrderStatusUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "BAD_REQUEST",
			"message": "Invalid request body.",
		})
		return
	}

	orderID := c.Param("order_id")
	req.OrderID = orderID

	_, apierr := h.serv.UpdateOrderStatus(ctx, req.OrderID, shipping.ShipmentStatus(req.Status))
	if apierr != nil {
		c.JSON(apierr.Status, gin.H{
			"error":   apierr.Code,
			"message": apierr.Message,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Order status updated successfully",
		"order":   orderID,
	})
}

// order returns

func (h *OrderHandler) ListAllReturns(c *gin.Context) {
	ctx := c.Request.Context()

	var filter order.ReturnFilter

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

	returnIDstr := c.Param("return_id")

	returnID, err := strconv.ParseInt(returnIDstr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "BAD_REQUEST",
			"message": "Invalid return ID.",
		})
		return
	}

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

	var req order.UpdateReturnRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "BAD_REQUEST",
			"message": "Invalid request body.",
		})
		return
	}

	id, err := strconv.ParseInt(returnID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "BAD_REQUEST",
			"message": "Invalid return ID.",
		})
		return
	}
	req.ReturnID = id

	if req.Status == "returned" {
		fmt.Println("Process return request called")
		_, apierr := h.serv.ProcessReturnRefundRequest(ctx, &req)
		if apierr != nil {
			c.JSON(apierr.Status, gin.H{
				"error":   apierr.Code,
				"message": apierr.Message,
			})
			return
		}
	} else {
		fmt.Println("Update return request status called")

		_, apierr := h.serv.UpdateReturnRequestStatus(ctx, &req)
		if apierr != nil {
			c.JSON(apierr.Status, gin.H{
				"error":   apierr.Code,
				"message": apierr.Message,
			})
			return
		}

	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Return request updated successfully",
	})
}

func (h *OrderHandler) GetReturnByID(c *gin.Context) {

}
