package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/tharunn0/E-Commerce-Go/internal/apperror"
	"github.com/tharunn0/E-Commerce-Go/internal/domain"
	"github.com/tharunn0/E-Commerce-Go/internal/utils"
	"go.uber.org/zap"
)

type OrderService struct {
	userRepo    domain.UserRepository
	productRepo domain.ProductRepository
	cartRepo    domain.CartRepository
	orderRepo   domain.OrderRepository
	log         *zap.Logger
}

func NewOrderService(userRepo domain.UserRepository, productRepo domain.ProductRepository, cartRepo domain.CartRepository, orderRepo domain.OrderRepository, log *zap.Logger) *OrderService {
	return &OrderService{
		userRepo:    userRepo,
		productRepo: productRepo,
		cartRepo:    cartRepo,
		orderRepo:   orderRepo,
		log:         log,
	}
}

// CHECKOUT SERVICES
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

	if userAddr.UserID != userID {

		return nil, nil, &apperror.APIError{
			Status:  http.StatusUnauthorized,
			Code:    "UNAUTHORIZED",
			Message: "You do not have access to this address.",
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

	var notEnoughStockErrors []domain.NotEnoughStockError

	// validate cart variant stocks
	for _, cartItem := range cart.Items {
		if cartVariantInfo[cartItem.ProductVariantID].Stock < cartItem.Quantity {
			notEnoughStockErrors = append(notEnoughStockErrors, domain.NotEnoughStockError{
				ProductVariantID: cartItem.ProductVariantID,
				SKU:              cartVariantInfo[cartItem.ProductVariantID].SKU,
				Quantity:         cartItem.Quantity,
				Stock:            cartVariantInfo[cartItem.ProductVariantID].Stock,
			})
		}
	}

	if len(notEnoughStockErrors) > 0 {
		return nil, notEnoughStockErrors, nil
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
// create order
func (s *OrderService) CreateOrderFromCart(ctx context.Context, req *domain.CreateOrderRequest) (*domain.CreateOrderResponse, []domain.NotEnoughStockError, *apperror.APIError) {
	// getuserid
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, nil, &apperror.APIError{
			Status:  http.StatusUnauthorized,
			Code:    "UNAUTHORIZED",
			Message: "You are not authorized to perform this action.",
		}
	}

	// validate address
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

	// validate address belongs to user
	if userAddr.UserID != userID {
		return nil, nil, &apperror.APIError{
			Status:  http.StatusUnauthorized,
			Code:    "UNAUTHORIZED",
			Message: "You do not have access to this address.",
		}
	}
	// validate delivery type
	if req.DeliveryType != domain.DeliveryTypeNormal && req.DeliveryType != domain.DeliveryTypeExpress {
		return nil, nil, &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "BAD_REQUEST",
			Message: "Invalid delivery type.",
		}
	}

	// validate cart and cart variant stocks
	cart, err := s.cartRepo.GetCartByUserID(ctx, userID)
	if err != nil {
		if err == apperror.ErrCartNotFound {
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

	cartVariantInfo, err := s.cartRepo.GetCartVariantInfo(ctx, userID)
	if err != nil {
		if err == apperror.ErrCartNotFound {
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

	var notEnoughStockErrors []domain.NotEnoughStockError

	// validate stock
	for _, cartItem := range cart.Items {
		if cartVariantInfo[cartItem.ProductVariantID].Stock < cartItem.Quantity {
			notEnoughStockErrors = append(notEnoughStockErrors, domain.NotEnoughStockError{
				ProductVariantID: cartItem.ProductVariantID,
				SKU:              cartVariantInfo[cartItem.ProductVariantID].SKU,
				Quantity:         cartItem.Quantity,
				Stock:            cartVariantInfo[cartItem.ProductVariantID].Stock,
			})
		}
	}

	if len(notEnoughStockErrors) > 0 {
		return nil, notEnoughStockErrors, nil
	}

	// create final order data

	orderID, err := utils.GeneratePublicOrderID()
	if err != nil {
		return nil, nil, &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "DB_ERROR",
			Message: "Failed to generate order ID.",
		}
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

	// total amount
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

	// create order && update stock
	if err := s.orderRepo.CreateOrder(ctx, orderData); err != nil {
		s.log.Error("Failed to create order", zap.Error(err))
		return nil, nil, &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "DB_ERROR",
			Message: "Failed to create order.",
		}
	}

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
		Status:                domain.OrderStatusPending,
		ShipmentStatus:        "PENDING",
		PaymentMethod:         "COD",
		PaymentStatus:         "PENDING",
		CreatedAt:             time.Now(),
	}

	s.log.Info("Order created successfully", zap.Any("order", resp))
	return resp, nil, nil
}

// get user orders
func (s *OrderService) GetOrders(ctx context.Context) ([]domain.OrderBaseResponse, *apperror.APIError) {
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

// ORDER ADMIN SERVICES

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
			Status:  http.StatusInternalServerError,
			Code:    "DB_ERROR",
			Message: "Failed to get orders.",
		}
	}
	return orders, nil
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
			Status:  http.StatusInternalServerError,
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
			Status:  http.StatusInternalServerError,
			Code:    "DB_ERROR",
			Message: "Failed to update order.",
		}
	}

	order.ShipmentStatus = "delivered"
	order.PaymentStatus = "paid"
	order.Status = "delivered"
	return order, nil
}
