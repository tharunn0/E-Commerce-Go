package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/tharunn0/E-Commerce-Go/internal/apperror"
	"github.com/tharunn0/E-Commerce-Go/internal/domain"
	"github.com/tharunn0/E-Commerce-Go/internal/infrastructure/payments"
	"github.com/tharunn0/E-Commerce-Go/internal/utils"

	Razorpay "github.com/razorpay/razorpay-go"
	"go.uber.org/zap"
)

type OrderService struct {
	userRepo    domain.UserRepository
	productRepo domain.ProductRepository
	cartRepo    domain.CartRepository
	orderRepo   domain.OrderRepository
	paymentRepo domain.PaymentRepository
	razorpay    *Razorpay.Client
	log         *zap.Logger
}

func NewOrderService(userRepo domain.UserRepository, productRepo domain.ProductRepository, cartRepo domain.CartRepository, orderRepo domain.OrderRepository, paymentRepo domain.PaymentRepository, razorpay *Razorpay.Client, log *zap.Logger) *OrderService {
	return &OrderService{
		userRepo:    userRepo,
		productRepo: productRepo,
		cartRepo:    cartRepo,
		orderRepo:   orderRepo,
		paymentRepo: paymentRepo,
		log:         log,
		razorpay:    razorpay,
	}
}

// CHECKOUT SERVICES
// ///////////////////////
// checkout cart
func (s *OrderService) CheckoutCart(ctx context.Context, req domain.CartCheckoutRequest) (*domain.CartCheckoutResponse, []domain.NotEnoughStockError, *apperror.APIError) {

	// validate delivery type
	if req.DeliveryType != domain.DeliveryTypeNormal && req.DeliveryType != domain.DeliveryTypeExpress {
		return nil, nil, &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "BAD_REQUEST",
			Message: "Invalid delivery type.",
		}
	}

	// get user id from context
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		s.log.Error("Failed to get user id from context", zap.Error(err))
		return nil, nil, &apperror.APIError{
			Status:  http.StatusUnauthorized,
			Code:    "UNAUTHORIZED",
			Message: "You are not authorized to perform this action.",
		}
	}

	// validate address
	userAddr, err := s.userRepo.GetUserAddressByID(ctx, req.AddressID)
	if err != nil {
		s.log.Error("Failed to get address", zap.Error(err))
		if err == apperror.ErrAddressNotFoundForUser {
			return nil, nil, &apperror.APIError{
				Status:  http.StatusNotFound,
				Code:    "NOT_FOUND",
				Message: apperror.ErrAddressNotFoundForUser.Error(),
			}
		}
		s.log.Error("Failed to get address", zap.Error(err))
		return nil, nil, &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "DB_ERROR",
			Message: "Failed to get address.",
		}
	}

	// validate address
	err = utils.ValidateUserAddress(userAddr, userID)
	if err != nil {
		return nil, nil, &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "BAD_REQUEST",
			Message: err.Error(),
		}
	}

	// get cart by user id
	cart, err := s.cartRepo.GetCartByUserID(ctx, userID)
	if err != nil {
		if err == apperror.ErrCartNotFound {
			s.log.Error("Failed to get cart", zap.Error(err))
			return nil, nil, &apperror.APIError{
				Status:  http.StatusNotFound,
				Code:    "NOT_FOUND",
				Message: apperror.ErrCartNotFound.Error(),
			}
		}
		s.log.Error("Failed to get cart", zap.Error(err))
		return nil, nil, &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "DB_ERROR",
			Message: "Failed to get cart.",
		}
	}

	if cart.Items == nil {
		return nil, nil, &apperror.APIError{
			Status:  http.StatusNotFound,
			Code:    "NOT_FOUND",
			Message: "Cart is empty.",
		}
	}

	// get cart variant

	cartVariantInfo, err := s.cartRepo.GetCartVariantInfo(ctx, userID)
	if err != nil {
		if err == apperror.ErrCartNotFound {
			s.log.Error("Failed to validate cart", zap.Error(err))
			return nil, nil, &apperror.APIError{
				Status:  http.StatusNotFound,
				Code:    "NOT_FOUND",
				Message: apperror.ErrCartNotFound.Error(),
			}
		}
		s.log.Error("Failed to validate cart", zap.Error(err))
		return nil, nil, &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "DB_ERROR",
			Message: "Failed to validate cart.",
		}
	}

	// validate not enough stock
	notEnoughStockError := utils.ValidateOrderItemsStock(cartVariantInfo, cart)
	if len(notEnoughStockError) > 0 {
		return nil, notEnoughStockError, nil
	}

	// shipping charge
	shippingAmount := domain.DeliveryTypeCharges[req.DeliveryType]

	// delivery time and date
	estimatedDeliveryTime, err := domain.GetDeliveryDays(userAddr.District)
	if err != nil {
		estimatedDeliveryTime = 7
		shippingAmount = 150
	}
	estimatedDeliveryDate, err := domain.CalculateDeliveryDate(userAddr.District)
	if err != nil {
		estimatedDeliveryDate = time.Now().AddDate(0, 0, estimatedDeliveryTime)
	}

	// final cart items price
	totalAmount := cart.CartTotalPrice + shippingAmount

	resp := &domain.CartCheckoutResponse{
		Cart:                  cart,
		ShippingCost:          shippingAmount,
		TotalAmount:           totalAmount,
		DeliveryType:          req.DeliveryType,
		Address:               userAddr,
		EstimatedDeliveryTime: fmt.Sprintf("%d days", estimatedDeliveryTime),
		EstimatedDeliveryDate: estimatedDeliveryDate.String(),
	}

	s.log.Info("Cart checkout successful", zap.Any("cart", cart), zap.Any("address", userAddr))

	return resp, nil, nil
}

// checkout product variant
func (s *OrderService) CheckoutProductVariant(ctx context.Context, req *domain.ProductVariantCheckoutRequest) (*domain.ProductVariantCheckoutResponse, *apperror.APIError) {
	activeonly := !utils.IsAdmin(ctx)
	// validate delivery type
	if req.DeliveryType != domain.DeliveryTypeNormal && req.DeliveryType != domain.DeliveryTypeExpress {
		return nil, &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "BAD_REQUEST",
			Message: "Invalid delivery type.",
		}
	}

	// get user id from context
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, &apperror.APIError{
			Status:  http.StatusUnauthorized,
			Code:    "UNAUTHORIZED",
			Message: "You are not authorized to perform this action.",
		}
	}

	// validate address exists
	userAddr, err := s.userRepo.GetUserAddressByID(ctx, req.AddressID)
	if err != nil {
		if err == apperror.ErrAddressNotFoundForUser {
			return nil, &apperror.APIError{
				Status:  http.StatusNotFound,
				Code:    "NOT_FOUND",
				Message: apperror.ErrAddressNotFoundForUser.Error(),
			}
		}
		s.log.Error("Failed to get address", zap.Error(err))
		return nil, &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "DB_ERROR",
			Message: "Failed to get address.",
		}
	}

	// validate address belongs to user
	if userAddr.UserID != userID {
		return nil, &apperror.APIError{
			Status:  http.StatusUnauthorized,
			Code:    "UNAUTHORIZED",
			Message: "You do not have access to this address.",
		}
	}

	productVariant, err := s.productRepo.GetProductVariantByID(ctx, req.ProductVariantID, activeonly)
	if err != nil {
		if err == apperror.ErrProductVariantNotFound {
			return nil, &apperror.APIError{
				Status:  http.StatusNotFound,
				Code:    "NOT_FOUND",
				Message: apperror.ErrProductVariantNotFound.Error(),
			}
		}
		s.log.Error("Failed to get product variant", zap.Error(err))
		return nil, &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "DB_ERROR",
			Message: "Failed to get product variant.",
		}
	}

	if productVariant.Stock < int(req.Quantity) {
		return nil, &apperror.APIError{
			Status:  http.StatusNotFound,
			Code:    "NOT_FOUND",
			Message: apperror.ErrNoStock.Error(),
		}
	}

	// final shipping charge
	shippingAmount := domain.DeliveryTypeCharges[req.DeliveryType]

	// final delivery time and date
	var totalAmount float64
	estimatedDeliveryTime, err := domain.GetDeliveryDays(userAddr.District)
	if err != nil {
		estimatedDeliveryTime = 7
		shippingAmount = 150
	}
	estimatedDeliveryDate, err := domain.CalculateDeliveryDate(userAddr.District)
	if err != nil {
		estimatedDeliveryDate = time.Now().AddDate(0, 0, estimatedDeliveryTime)
	}

	// final product variant price
	if productVariant.SalePrice == nil {
		totalAmount = productVariant.OriginalPrice + shippingAmount
	} else {
		totalAmount = *productVariant.SalePrice + shippingAmount
	}

	resp := &domain.ProductVariantCheckoutResponse{
		ProductVariant:        productVariant,
		ShippingCost:          shippingAmount,
		TotalAmount:           totalAmount,
		DeliveryType:          req.DeliveryType,
		Address:               userAddr,
		EstimatedDeliveryTime: fmt.Sprintf("%d days", estimatedDeliveryTime),
		EstimatedDeliveryDate: estimatedDeliveryDate.String(),
	}

	return resp, nil
}

// ORDER SERVICES
// ////////////////
// create order
func (s *OrderService) CreateOrderFromCart(ctx context.Context, req *domain.CreateOrderRequest) (*domain.CreateOrderResponse, []domain.NotEnoughStockError, *apperror.APIError) {

	// 1. validate req
	if req.DeliveryType != domain.DeliveryTypeNormal && req.DeliveryType != domain.DeliveryTypeExpress {
		return nil, nil, &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "BAD_REQUEST",
			Message: "Invalid delivery type.",
		}
	}
	if req.AddressID == 0 {
		return nil, nil, &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "BAD_REQUEST",
			Message: "Invalid address.",
		}
	}

	// 2. extract user id
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, nil, &apperror.APIError{
			Status:  http.StatusUnauthorized,
			Code:    "UNAUTHORIZED",
			Message: "You are not authorized to perform this action.",
		}
	}

	// 3. get address
	userAddr, err := s.userRepo.GetUserAddressByID(ctx, req.AddressID)
	if err != nil {
		if err == apperror.ErrAddressNotFoundForUser {
			return nil, nil, &apperror.APIError{
				Status:  http.StatusNotFound,
				Code:    "NOT_FOUND",
				Message: apperror.ErrAddressNotFoundForUser.Error(),
			}
		}
		s.log.Error("Failed to get address", zap.Error(err))
		return nil, nil, &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "DB_ERROR",
			Message: "Failed to get address.",
		}
	}

	// 4. validate address belongs to user
	err = utils.ValidateUserAddress(userAddr, userID)
	if err != nil {
		return nil, nil, &apperror.APIError{
			Status:  http.StatusUnauthorized,
			Code:    "UNAUTHORIZED",
			Message: "You do not have access to this address.",
		}
	}

	// 5. fetch cart with userId
	cart, err := s.cartRepo.GetCartByUserID(ctx, userID)
	if err != nil {
		if err == apperror.ErrCartNotFound {
			return nil, nil, &apperror.APIError{
				Status:  http.StatusNotFound,
				Code:    "NOT_FOUND",
				Message: apperror.ErrCartNotFound.Error(),
			}
		}
		s.log.Error("Failed to fetch cart", zap.Error(err))
		return nil, nil, &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "DB_ERROR",
			Message: "Failed to fetch cart.",
		}
	}

	if cart.Items == nil {
		return nil, nil, &apperror.APIError{
			Status:  http.StatusNotFound,
			Code:    "NOT_FOUND",
			Message: "Cart is empty.",
		}
	}

	// 6. fetch variant info
	cartVariantInfo, err := s.cartRepo.GetCartVariantInfo(ctx, userID)
	if err != nil {
		if err == apperror.ErrCartNotFound {
			return nil, nil, &apperror.APIError{
				Status:  http.StatusNotFound,
				Code:    "NOT_FOUND",
				Message: apperror.ErrCartNotFound.Error(),
			}
		}
		s.log.Error("Failed to fetch variant info", zap.Error(err))
		return nil, nil, &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "DB_ERROR",
			Message: "Failed to fetch variant info.",
		}
	}

	// 7. validate stock
	notEnoughStockErrors := utils.ValidateOrderItemsStock(cartVariantInfo, cart)
	if len(notEnoughStockErrors) > 0 {
		return nil, notEnoughStockErrors, nil
	}

	// 8. generate public order id
	orderID, err := utils.GeneratePublicOrderID()
	if err != nil {
		return nil, nil, &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Failed to generate order ID.",
		}
	}

	// 9. calculate shipping charge
	shippingAmount := domain.DeliveryTypeCharges[req.DeliveryType]

	// 10. calculate delivery time and date
	estimatedDeliveryTime, err := domain.GetDeliveryDays(userAddr.District)
	if err != nil {
		estimatedDeliveryTime = 7
		shippingAmount = 150
	}
	estimatedDeliveryDate, err := domain.CalculateDeliveryDate(userAddr.District)
	if err != nil {
		estimatedDeliveryDate = time.Now().AddDate(0, 0, estimatedDeliveryTime)
	}

	// 11. calculate total amount
	totalAmount := cart.CartTotalPrice + shippingAmount

	orderData := &domain.CreateOrderData{
		UserID:                userID,
		OrderID:               orderID,
		TotalAmount:           totalAmount,
		TaxAmount:             0,
		ShippingAddressID:     req.AddressID,
		BillingAddressID:      req.AddressID,
		DeliveryType:          strings.ToLower(string(req.DeliveryType)),
		EstimatedDeliveryDate: estimatedDeliveryDate,
		OrderSource:           "cart",
	}

	// 12. create order items
	var items []domain.OrderItem

	for _, cartItem := range cart.Items {
		var item domain.OrderItem
		item.ProductVariantID = cartItem.ProductVariantID
		item.ProductName = cartVariantInfo[cartItem.ProductVariantID].ProductName
		item.SKU = cartVariantInfo[cartItem.ProductVariantID].SKU
		item.Quantity = cartItem.Quantity
		item.TotalPrice = cartItem.TotalPrice
		if cartItem.SalePrice != nil {
			item.UnitPrice = *cartItem.SalePrice
		} else {
			item.UnitPrice = cartItem.OriginalPrice
		}
		items = append(items, item)
	}
	orderData.Items = items

	// 13. validate payment gateway

	gateway := payments.GetPaymentGateway(req.PaymentMethod, s.razorpay)
	if gateway == nil {
		return nil, nil, &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "BAD_REQUEST",
			Message: "Invalid payment method.",
		}
	}

	if req.PaymentMethod == domain.PaymentMethodCOD {
		orderData.Status = domain.OrderStatusConfirmed
	} else {
		orderData.Status = domain.OrderStatusPending
	}

	// 14. create order && update stock
	if err := s.orderRepo.CreateOrder(ctx, orderData); err != nil {
		s.log.Error("Failed to create order", zap.Error(err))
		return nil, nil, &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "DB_ERROR",
			Message: "Failed to create order.",
		}
	}

	// 15. create payment and payment response
	paymentResp, err := gateway.CreatePayment(ctx, domain.PaymentRequest{
		OrderID:  orderID,
		UserID:   userID,
		Amount:   int64(totalAmount) / 100,
		Currency: "INR",
	})
	if err != nil {
		s.log.Error("Failed to initialize payment", zap.Error(err))
		return nil, nil, &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "DB_ERROR",
			Message: "Failed to initialize payment.",
		}
	}

	payment := &domain.Payment{
		OrderID:    orderID,
		UserID:     userID,
		Amount:     int64(totalAmount) * 100,
		Currency:   "INR",
		Provider:   req.PaymentMethod,
		Status:     paymentResp.Status,
		GatewayRef: paymentResp.GatewayRef,
		PaymentURL: paymentResp.PaymentURL,
	}

	if err := s.paymentRepo.CreatePayment(ctx, payment); err != nil {
		s.log.Error("Failed to create payment", zap.Error(err))
		return nil, nil, &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "DB_ERROR",
			Message: "Failed to create payment.",
		}
	}

	s.log.Info("Payment initialized", zap.Any("payment", payment))

	// return response
	resp := &domain.CreateOrderResponse{
		OrderID:               orderID,
		Items:                 items,
		Subtotal:              cart.CartTotalPrice,
		TaxAmount:             0,
		ShippingCost:          shippingAmount,
		TotalAmount:           totalAmount,
		Currency:              "INR",
		ShippingAddressID:     req.AddressID,
		ShippingAddress:       userAddr,
		BillingAddressID:      req.AddressID,
		DeliveryType:          req.DeliveryType,
		EstimatedDeliveryTime: fmt.Sprintf("%d days", estimatedDeliveryTime),
		EstimatedDeliveryDate: estimatedDeliveryDate.String(),
		Status:                orderData.Status,
		ShipmentStatus:        string(domain.ShipmentStatusPending),
		PaymentMethod:         string(req.PaymentMethod),
		PaymentStatus:         string(paymentResp.Status),
		Payment:               payment,
		CreatedAt:             time.Now(),
	}

	if req.PaymentMethod == domain.PaymentMethodCOD {
		resp.Payment = nil
	}

	s.log.Info("Order created successfully", zap.Any("order", resp))
	return resp, nil, nil
}

func (s *OrderService) UpdateOrderStatusOnPayment(ctx context.Context, status domain.PaymentStatus, orderID string) *apperror.APIError {

	// 1. validate payment status
	if status != domain.PaymentStatusCompleted && status != domain.PaymentStatusFailed {
		s.log.Warn("Invalid payment status", zap.String("status", string(status)))
		return &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "BAD_REQUEST",
			Message: "Invalid payment status.",
		}
	}

	// 2. update payment status

	if err := s.paymentRepo.UpdatePaymentStatus(ctx, orderID, status); err != nil {
		s.log.Error("Failed to update payment status", zap.Error(err))
		return &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "DB_ERROR",
			Message: "Failed to update payment status.",
		}
	}

	// 3. update order status
	err := s.orderRepo.UpdateOrderStatusOnPayment(ctx, orderID, status)
	if err != nil {
		s.log.Error("Failed to update order status", zap.Error(err))
		return &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "DB_ERROR",
			Message: "Failed to update order status.",
		}
	}

	// 4. return response

	return nil

}

// get user orders
func (s *OrderService) GetUserOrders(ctx context.Context) ([]domain.OrderBaseResponse, *apperror.APIError) {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		s.log.Error("Failed to get user ID", zap.Error(err))
		return nil, &apperror.APIError{
			Status:  http.StatusUnauthorized,
			Code:    "UNAUTHORIZED",
			Message: "You are not authorized to perform this action.",
		}
	}
	orders, err := s.orderRepo.GetUserOrders(ctx, userID)
	if err != nil {
		s.log.Error("Failed to get orders", zap.Error(err))
		return nil, &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "DB_ERROR",
			Message: "Failed to get orders.",
		}
	}

	return orders, nil
}

func (s *OrderService) GetOrderByID(ctx context.Context, orderID string) (*domain.OrderResponse, *apperror.APIError) {
	order, err := s.orderRepo.GetUserOrderByID(ctx, orderID)
	if err != nil {
		s.log.Error("Failed to get order", zap.Error(err))
		if err == apperror.ErrOrderNotFound {
			return nil, &apperror.APIError{
				Status:  http.StatusNotFound,
				Code:    "NOT_FOUND",
				Message: apperror.ErrOrderNotFound.Error(),
			}
		}
		return nil, &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "DB_ERROR",
			Message: "Failed to get order.",
		}
	}
	order.ShippingCost = domain.DeliveryTypeCharges[order.DeliveryType]
	order.Subtotal = order.TotalAmount - order.TaxAmount - order.ShippingCost
	return order, nil
}

// cancel order item
func (s *OrderService) CancelOrderItem(ctx context.Context, orderID string, variantID int64) (*domain.OrderResponse, *apperror.APIError) {

	order, err := s.orderRepo.GetUserOrderByID(ctx, orderID)
	if err != nil {
		s.log.Error("Failed to get order", zap.Error(err))
		if err == apperror.ErrOrderNotFound {
			return nil, &apperror.APIError{
				Status:  http.StatusNotFound,
				Code:    "NOT_FOUND",
				Message: apperror.ErrOrderNotFound.Error(),
			}
		}
		return nil, &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "DB_ERROR",
			Message: "Failed to get order.",
		}
	}

	if order.Status == domain.OrderStatusDelivered {
		return nil, &apperror.APIError{
			Status:  http.StatusNotFound,
			Code:    "NOT_FOUND",
			Message: "Order is already delivered.Choose to return the order.",
		}
	}

	if order.Status == domain.OrderStatusCancelled {
		return nil, &apperror.APIError{
			Status:  http.StatusNotFound,
			Code:    "NOT_FOUND",
			Message: "Order is already cancelled.",
		}
	}

	variantFound := false
	for _, item := range order.Items {
		if item.ProductVariantID == variantID {
			variantFound = true
			break
		}
	}

	if !variantFound {
		return nil, &apperror.APIError{
			Status:  http.StatusNotFound,
			Code:    "NOT_FOUND",
			Message: "Variant not found in order.",
		}
	}

	err = s.orderRepo.CancelOrderItem(ctx, orderID, variantID)
	if err != nil {
		s.log.Error("Failed to cancel order item", zap.Error(err))
		if err == apperror.ErrOrderNotFound {
			return nil, &apperror.APIError{
				Status:  http.StatusNotFound,
				Code:    "NOT_FOUND",
				Message: apperror.ErrOrderNotFound.Error(),
			}
		}
		return nil, &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "DB_ERROR",
			Message: "Failed to cancel order item.",
		}
	}

	UpdatedOrder, err := s.orderRepo.GetUserOrderByID(ctx, orderID)
	if err != nil {
		s.log.Error("Failed to get order", zap.Error(err))
		if err == apperror.ErrOrderNotFound {
			return nil, &apperror.APIError{
				Status:  http.StatusNotFound,
				Code:    "NOT_FOUND",
				Message: apperror.ErrOrderNotFound.Error(),
			}
		}
		return nil, &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "DB_ERROR",
			Message: "Failed to get order.",
		}
	}

	return UpdatedOrder, nil
}

// cancel order
func (s *OrderService) CancelOrder(ctx context.Context, req *domain.CancelOrderRequest) (*domain.OrderResponse, *apperror.APIError) {
	if req.OrderID == "" {
		return nil, &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_ORDER_ID",
			Message: "Provide a valid order ID.",
		}
	}

	if req.Reason == "" || len(req.Reason) < 5 {
		return nil, &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_REASON",
			Message: "Provide a valid reason with at least 5 characters.",
		}
	}

	order, err := s.orderRepo.GetUserOrderByID(ctx, req.OrderID)
	if err != nil {
		s.log.Error("Failed to get order", zap.Error(err))
		if err == apperror.ErrOrderNotFound {
			return nil, &apperror.APIError{
				Status:  http.StatusNotFound,
				Code:    "NOT_FOUND",
				Message: apperror.ErrOrderNotFound.Error(),
			}
		}
		return nil, &apperror.APIError{
			Status:  http.StatusConflict,
			Code:    "DB_ERROR",
			Message: "Failed to get order.",
		}
	}

	statusStr := strings.ToUpper(string(order.Status))

	switch statusStr {
	case string(domain.OrderStatusCancelled):
		return nil, &apperror.APIError{
			Status:  http.StatusNotFound,
			Code:    "NOT_FOUND",
			Message: "Order is already cancelled.",
		}
	case string(domain.OrderStatusDelivered):
		return nil, &apperror.APIError{
			Status:  http.StatusNotFound,
			Code:    "NOT_FOUND",
			Message: "Order is already delivered.Choose to return the order.",
		}
	case string(domain.OrderStatusPending):
		return nil, &apperror.APIError{
			Status:  http.StatusNotFound,
			Code:    "NOT_FOUND",
			Message: "Order has not been confirmed yet.",
		}
	}

	err = s.orderRepo.CancelOrder(ctx, req.OrderID, req.Reason)
	if err != nil {
		s.log.Error("Failed to cancel order", zap.Error(err))
		return nil, &apperror.APIError{
			Status:  http.StatusConflict,
			Code:    "DB_ERROR",
			Message: "Failed to cancel order.",
		}
	}

	s.log.Info("Order cancelled successfully", zap.String("order_id", req.OrderID))

	if order.PaymentMethod == string(domain.PaymentMethodRazorpay) {
		// err := s.paymentRepo.CancelOrderPayment(ctx, req.OrderID)
		// if err != nil {
		// 	s.log.Error("Failed to cancel order payment", zap.Error(err))
		// 	return nil, &apperror.APIError{
		// 		Status:  http.StatusInternalServerError,
		// 		Code:    "DB_ERROR",
		// 		Message: "Failed to cancel order payment.",
		// 	}
		// }
	}

	return nil, nil
}

// ORDER ADMIN SERVICES
// ///////////////////////
// list all orders
func (s *OrderService) ListAllOrders(ctx context.Context, filter *domain.OrderFilter) ([]domain.OrderBaseResponse, *apperror.APIError) {

	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Sort == nil {
		sort := "DESC"
		filter.Sort = &sort
	}
	if filter.OrderBy == nil {
		createdAt := "o.created_at"
		filter.OrderBy = &createdAt
	}

	orders, err := s.orderRepo.ListAllOrders(ctx, filter)
	if err != nil {
		s.log.Error("Failed to get orders", zap.Error(err))
		return nil, &apperror.APIError{
			Status:  http.StatusConflict,
			Code:    "DB_ERROR",
			Message: "Failed to get orders.",
		}
	}
	return orders, nil
}

// update order status
func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderID string, status domain.ShipmentStatus) (*domain.OrderResponse, *apperror.APIError) {

	// check if order exists

	order, err := s.orderRepo.GetUserOrderByID(ctx, orderID)
	if err != nil {
		s.log.Error("Failed to get order", zap.Error(err))
		if err == apperror.ErrOrderNotFound {
			return nil, &apperror.APIError{
				Status:  http.StatusNotFound,
				Code:    "NOT_FOUND",
				Message: apperror.ErrOrderNotFound.Error(),
			}
		}
		return nil, &apperror.APIError{
			Status:  http.StatusConflict,
			Code:    "DB_ERROR",
			Message: "Failed to get order.",
		}
	}

	fmt.Println("order status", order.Status, domain.OrderStatusFailed)
	fmt.Println("new status", status)

	if status != "shipped" && status != "delivered" {
		return nil, &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_STATUS",
			Message: "Invalid status.",
		}
	}

	strStatus := strings.ToUpper(string(order.Status)) //
	strShipmentStatus := order.ShipmentStatus

	// check if status is valid for the order
	switch strStatus {
	case string(domain.OrderStatusCancelled):
		return nil, &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_STATUS",
			Message: "Order is already cancelled.",
		}
	case string(domain.OrderStatusDelivered):
		return nil, &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_STATUS",
			Message: "Order is already delivered.",
		}
	case string(domain.OrderStatusPending):
		return nil, &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_STATUS",
			Message: "Order has not been confirmed.",
		}
	case string(domain.OrderStatusFailed):
		return nil, &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_STATUS",
			Message: "Order has not been confirmed.",
		}
	}

	switch strStatus {
	case string(domain.ShipmentStatusShipped):
		if order.ShipmentStatus == string(domain.ShipmentStatusDelivered) {
			return nil, &apperror.APIError{
				Status:  http.StatusNotFound,
				Code:    "NOT_FOUND",
				Message: "Order is already shipped.",
			}
		}
	case string(domain.ShipmentStatusDelivered):
		if order.ShipmentStatus == string(domain.ShipmentStatusDelivered) {
			return nil, &apperror.APIError{
				Status:  http.StatusNotFound,
				Code:    "NOT_FOUND",
				Message: "Order is already delivered.",
			}
		}
	}

	if string(status) == "delivered" && strShipmentStatus != "shipped" {
		return nil, &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_STATUS",
			Message: "Order is not shipped.",
		}
	}

	isCOD := false
	if order.PaymentMethod == "COD" {
		isCOD = true
	}

	// update order status
	if err := s.orderRepo.UpdateShipmentStatus(ctx, orderID, status, isCOD); err != nil {
		s.log.Error("Failed to update order status", zap.Error(err))
		return nil, &apperror.APIError{
			Status:  http.StatusConflict,
			Code:    "DB_ERROR",
			Message: "Failed to update order status.",
		}
	}

	//

	return nil, nil
}

// ship order
func (s *OrderService) ShipOrder(ctx context.Context, orderID string) (*domain.OrderResponse, *apperror.APIError) {
	order, err := s.orderRepo.GetUserOrderByID(ctx, orderID)
	if err != nil {
		s.log.Error("Failed to get order", zap.Error(err))
		if err == apperror.ErrOrderNotFound {
			return nil, &apperror.APIError{
				Status:  http.StatusNotFound,
				Code:    "NOT_FOUND",
				Message: apperror.ErrOrderNotFound.Error(),
			}
		}
		return nil, &apperror.APIError{
			Status:  http.StatusConflict,
			Code:    "DB_ERROR",
			Message: "Failed to get order.",
		}
	}

	carrier := domain.SelectRandomCarrier()
	trackingID := domain.GenerateTrackingID(carrier)

	shipmentData := &domain.ShipmentData{
		Carrier:    carrier,
		TrackingID: trackingID,
		ShippedAt:  time.Now(),
	}

	if err := s.orderRepo.ShipOrder(ctx, orderID, shipmentData); err != nil {
		s.log.Error("Failed to ship order", zap.Error(err))
		return nil, &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "DB_ERROR",
			Message: "Failed to update order.",
		}
	}

	order.ShippingCost = domain.DeliveryTypeCharges[order.DeliveryType]
	order.Subtotal = order.TotalAmount - order.TaxAmount - order.ShippingCost

	fmt.Println("delivery type", order.DeliveryType)
	fmt.Println("shipping cost", order.ShippingCost)
	fmt.Println("subtotal", order.Subtotal)

	order.ShipmentStatus = "shipped"
	order.ShipmentCarrier = string(carrier)
	order.TrackingNumber = trackingID
	return order, nil
}

// deliver order
func (s *OrderService) DeliverOrder(ctx context.Context, orderID string) (*domain.OrderResponse, *apperror.APIError) {
	order, err := s.orderRepo.GetUserOrderByID(ctx, orderID)
	if err != nil {
		s.log.Error("Failed to get order", zap.Error(err))
		if err == apperror.ErrOrderNotFound {
			return nil, &apperror.APIError{
				Status:  http.StatusNotFound,
				Code:    "NOT_FOUND",
				Message: apperror.ErrOrderNotFound.Error(),
			}
		}
		return nil, &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "DB_ERROR",
			Message: "Failed to get order.",
		}
	}

	if err := s.orderRepo.DeliverOrder(ctx, orderID); err != nil {
		s.log.Error("Failed to deliver order", zap.Error(err))
		return nil, &apperror.APIError{
			Status:  http.StatusConflict,
			Code:    "DB_ERROR",
			Message: "Failed to update order.",
		}
	}

	order.ShipmentStatus = "delivered"
	order.PaymentStatus = "paid"
	order.Status = "delivered"
	return order, nil
}

func (s *OrderService) ReturnOrderItemRequest(ctx context.Context, req *domain.ReturnOrderItemRequest) (*domain.OrderResponse, *apperror.APIError) {

	if req.OrderID == "" || req.VariantID <= 0 {
		return nil, &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "BAD_REQUEST",
			Message: "Invalid request.",
		}
	}

	if req.Reason == "" || len(req.Reason) < 10 {
		return nil, &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "BAD_REQUEST",
			Message: "Please provide a valid return reason.",
		}
	}

	order, err := s.orderRepo.GetUserOrderByID(ctx, req.OrderID)
	if err != nil {
		s.log.Error("Failed to get order", zap.Error(err))
		if err == apperror.ErrOrderNotFound {
			return nil, &apperror.APIError{
				Status:  http.StatusNotFound,
				Code:    "NOT_FOUND",
				Message: apperror.ErrOrderNotFound.Error(),
			}
		}
		return nil, &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "DB_ERROR",
			Message: "Failed to get order.",
		}
	}

	strStatus := strings.ToUpper(string(order.Status))

	if strStatus != string(domain.OrderStatusDelivered) {
		return nil, &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "BAD_REQUEST",
			Message: "Order can be only returned after delivery.",
		}
	}

	return nil, nil
}
