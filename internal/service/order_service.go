package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/tharunn0/E-Commerce-Go/internal/apperror"
	"github.com/tharunn0/E-Commerce-Go/internal/config"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/cart"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/discount"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/order"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/payment"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/product"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/promotion"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/shipping"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/user"
	"github.com/tharunn0/E-Commerce-Go/internal/infrastructure/payments"
	"github.com/tharunn0/E-Commerce-Go/internal/utils"

	Razorpay "github.com/razorpay/razorpay-go"
	"go.uber.org/zap"
)

type OrderService struct {
	userRepo    user.UserRepository
	productRepo product.ProductRepository
	cartRepo    cart.CartRepository
	orderRepo   order.OrderRepository
	paymentRepo payment.PaymentRepository
	offerRepo   promotion.OfferRepository
	couponRepo  promotion.CouponRepository
	razorpay    *Razorpay.Client
	cfg         config.OrderSettings
	log         *zap.Logger
}

func NewOrderService(
	userRepo user.UserRepository,
	productRepo product.ProductRepository,
	cartRepo cart.CartRepository,
	orderRepo order.OrderRepository,
	paymentRepo payment.PaymentRepository,
	offerRepo promotion.OfferRepository,
	couponRepo promotion.CouponRepository,
	razorpay *Razorpay.Client,
	cfg config.OrderSettings,
	log *zap.Logger,
) *OrderService {
	return &OrderService{
		userRepo:    userRepo,
		productRepo: productRepo,
		cartRepo:    cartRepo,
		orderRepo:   orderRepo,
		paymentRepo: paymentRepo,
		offerRepo:   offerRepo,
		couponRepo:  couponRepo,
		log:         log,
		razorpay:    razorpay,
		cfg:         cfg,
	}
}

// CHECKOUT SERVICES
// ///////////////////////
// checkout cart
func (s *OrderService) CheckoutCart(ctx context.Context, req order.CartCheckoutRequest) (*order.CartCheckoutResponse, []order.NotEnoughStockError, *apperror.APIError) {

	// validate delivery type
	if req.DeliveryType != shipping.DeliveryTypeNormal && req.DeliveryType != shipping.DeliveryTypeExpress {
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
	cartObj, err := s.cartRepo.GetCartByUserID(ctx, userID)
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

	if cartObj.Items == nil {
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
	notEnoughStockError := utils.ValidateOrderItemsStock(cartVariantInfo, cartObj)
	if len(notEnoughStockError) > 0 {
		return nil, notEnoughStockError, nil
	}

	var productIDs []int64
	var categoryIDs []int64
	for _, item := range cartObj.Items {
		productIDs = append(productIDs, item.ProductID)
		categoryIDs = append(categoryIDs, item.CategoryID)
	}

	offers, err := s.offerRepo.GetAllActiveOffers(ctx, productIDs, categoryIDs)
	if err != nil {
		s.log.Error("failed to get offers", zap.String("function", "GetAllActiveOffers"), zap.Error(err))
		return nil, nil, &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to get offers.",
		}
	}

	discount.ApplyDiscounts(cartObj, offers)

	// shipping charge
	shippingAmount := shipping.DeliveryTypeCharges[req.DeliveryType]

	// delivery time and date
	estimatedDeliveryTime, err := shipping.GetDeliveryDays(userAddr.District)
	if err != nil {
		estimatedDeliveryTime = 7
		shippingAmount = 150
	}
	estimatedDeliveryDate, err := shipping.CalculateDeliveryDate(userAddr.District)
	if err != nil {
		estimatedDeliveryDate = time.Now().AddDate(0, 0, estimatedDeliveryTime)
	}

	// final cart items price
	totalAmount := cartObj.CartTotalPrice + shippingAmount

	resp := &order.CartCheckoutResponse{
		Cart:                  cartObj,
		ShippingCost:          shippingAmount,
		TotalAmount:           totalAmount,
		DeliveryType:          req.DeliveryType,
		Address:               userAddr,
		EstimatedDeliveryTime: fmt.Sprintf("%d days", estimatedDeliveryTime),
		EstimatedDeliveryDate: estimatedDeliveryDate.String(),
	}

	s.log.Info("Cart checkout successful", zap.Any("cart", cartObj), zap.Any("address", userAddr))

	return resp, nil, nil
}

// checkout product variant
func (s *OrderService) CheckoutProductVariant(ctx context.Context, req *order.ProductVariantCheckoutRequest) (*order.ProductVariantCheckoutResponse, *apperror.APIError) {
	activeonly := !utils.IsAdmin(ctx)
	// validate delivery type
	if req.DeliveryType != shipping.DeliveryTypeNormal && req.DeliveryType != shipping.DeliveryTypeExpress {
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
	shippingAmount := shipping.DeliveryTypeCharges[req.DeliveryType]

	// final delivery time and date
	var totalAmount float64
	estimatedDeliveryTime, err := shipping.GetDeliveryDays(userAddr.District)
	if err != nil {
		estimatedDeliveryTime = 7
		shippingAmount = 150
	}
	estimatedDeliveryDate, err := shipping.CalculateDeliveryDate(userAddr.District)
	if err != nil {
		estimatedDeliveryDate = time.Now().AddDate(0, 0, estimatedDeliveryTime)
	}

	// final product variant price
	if productVariant.SalePrice == nil {
		totalAmount = productVariant.OriginalPrice + shippingAmount
	} else {
		totalAmount = *productVariant.SalePrice + shippingAmount
	}

	resp := &order.ProductVariantCheckoutResponse{
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

// create order
func (s *OrderService) CreateOrderFromCart(ctx context.Context, req *order.CreateOrderRequest) (*order.CreateOrderResponse, []order.NotEnoughStockError, *apperror.APIError) {

	// 1. validate req
	if err := req.Validate(); err != nil {
		return nil, nil, apperror.New(http.StatusBadRequest, "BAD_REQUEST", err.Error())
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

	fmt.Println("Checking user ID: ", userID)

	// 3. get address
	userAddr, err := s.userRepo.GetUserAddressByID(ctx, req.AddressID)
	if err != nil {
		if err == apperror.ErrAddressNotFoundForUser {
			return nil, nil, apperror.New(http.StatusNotFound, "NOT_FOUND", apperror.ErrAddressNotFoundForUser.Error())
		}
		s.log.Error("Failed to get address", zap.Error(err))
		return nil, nil, apperror.New(http.StatusInternalServerError, "DB_ERROR", "Failed to get address.")
	}

	// 4. validate address belongs to user
	err = utils.ValidateUserAddress(userAddr, userID)
	if err != nil {
		return nil, nil, apperror.New(http.StatusUnauthorized, "UNAUTHORIZED", "You do not have access to this address.")
	}

	// 5. fetch cart with userId
	cartObj, err := s.cartRepo.GetCartByUserID(ctx, userID)
	if err != nil {
		if err == apperror.ErrCartNotFound {
			return nil, nil, apperror.New(http.StatusNotFound, "NOT_FOUND", apperror.ErrCartNotFound.Error())
		}
		s.log.Error("Failed to fetch cart", zap.Error(err))
		return nil, nil, apperror.New(http.StatusInternalServerError, "DB_ERROR", "Failed to fetch cart.")
	}

	if cartObj.Items == nil {
		return nil, nil, apperror.New(http.StatusNotFound, "NOT_FOUND", "Cart is empty.")
	}

	if cartObj.CartTotalPrice > float64(s.cfg.MaxOrderAmount) {
		return nil, nil, apperror.New(http.StatusUnauthorized, "UNAUTHORIZED", fmt.Sprintf("Order amount exceeded. Should be less than %d.", s.cfg.MaxOrderAmount))
	}

	// 6. fetch variant info
	cartVariantInfo, err := s.cartRepo.GetCartVariantInfo(ctx, userID)
	if err != nil {
		if err == apperror.ErrCartNotFound {
			return nil, nil, apperror.New(http.StatusNotFound, "NOT_FOUND", apperror.ErrCartNotFound.Error())
		}
		s.log.Error("Failed to fetch variant info", zap.Error(err))
		return nil, nil, apperror.New(http.StatusInternalServerError, "DB_ERROR", "Failed to fetch variant info.")
	}

	// 7. validate stock
	notEnoughStockErrors := utils.ValidateOrderItemsStock(cartVariantInfo, cartObj)
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

	var productIds, categoryIds []int64

	for _, item := range cartObj.Items {
		productIds = append(productIds, item.ProductID)
		categoryIds = append(categoryIds, item.CategoryID)
	}

	// apply existing offers

	offers, err := s.offerRepo.GetAllActiveOffers(ctx, productIds, categoryIds)
	if err != nil {
		s.log.Error("Failed to get offers", zap.Error(err))
		return nil, nil, apperror.New(http.StatusInternalServerError, "DB_ERROR", "Failed to get offers.")
	}

	if len(offers) > 0 {
		discount.ApplyDiscounts(cartObj, offers)
	}

	// apply coupons

	if req.CouponCode != "" {
		coupon, err := s.couponRepo.FetchCoupon(ctx, req.CouponCode, userID)
		if err != nil {
			s.log.Error("Failed to get coupon", zap.Error(err))
			if err == apperror.ErrCouponNotFound {
				return nil, nil, apperror.New(http.StatusNotFound, "NOT_FOUND", apperror.ErrCouponNotFound.Error())
			}
			return nil, nil, apperror.New(http.StatusInternalServerError, "DB_ERROR", "Failed to get coupon.")
		}

		err = promotion.ValidateCoupon(coupon, time.Now())
		if err != nil {
			return nil, nil, apperror.New(http.StatusBadRequest, "BAD_REQUEST", err.Error())
		}

		err = cart.ApplyCouponToCart(cartObj, coupon)
		if err != nil {
			return nil, nil, apperror.New(http.StatusBadRequest, "BAD_REQUEST", err.Error())
		}
	}

	// 9. calculate shipping charge
	shippingAmount := shipping.DeliveryTypeCharges[req.DeliveryType]

	// 10. calculate delivery time and date
	estimatedDeliveryTime, err := shipping.GetDeliveryDays(userAddr.District)
	if err != nil {
		estimatedDeliveryTime = 7
		shippingAmount = 150
	}
	estimatedDeliveryDate, err := shipping.CalculateDeliveryDate(userAddr.District)
	if err != nil {
		estimatedDeliveryDate = time.Now().AddDate(0, 0, estimatedDeliveryTime)
	}

	// 11. calculate total amount
	totalAmount := cartObj.CartTotalPrice + shippingAmount

	if totalAmount > float64(s.cfg.MaxCodOrderAmount) && req.PaymentMethod == payment.PaymentMethodCOD {
		return nil, nil, apperror.New(http.StatusBadRequest, "BAD_REQUEST", fmt.Sprintf("Order amount exceeded. Should be less than %d for COD.", s.cfg.MaxCodOrderAmount))
	}

	orderData := &order.CreateOrderData{
		UserID:                userID,
		OrderID:               orderID,
		TotalAmount:           totalAmount,
		TaxAmount:             0,
		ShippingAddressID:     req.AddressID,
		BillingAddressID:      req.AddressID,
		DeliveryType:          strings.ToLower(string(req.DeliveryType)),
		EstimatedDeliveryDate: estimatedDeliveryDate,
		OrderSource:           "cart",
		CouponData:            cartObj.CouponData,
	}

	// 12. create order items
	orderData.Items = utils.CreateOrderItems(cartObj, cartVariantInfo)

	// 13. validate payment gateway
	gateway := payments.GetPaymentGateway(req.PaymentMethod, s.razorpay)
	if gateway == nil {
		return nil, nil, apperror.New(http.StatusBadRequest, "BAD_REQUEST", "Invalid payment method.")
	}

	switch req.PaymentMethod {
	case payment.PaymentMethodCOD:
		orderData.Status = order.OrderStatusConfirmed
	case payment.PaymentMethodRazorpay:
		orderData.Status = order.OrderStatusPending
	case payment.PaymentMethodWallet:
		// fetch wallet and check if amount is enough
		wallet, err := s.userRepo.GetWallet(ctx, userID)
		if err != nil {
			return nil, nil, apperror.New(http.StatusInternalServerError, "DB_ERROR", "Failed to get wallet.")
		}
		if wallet.Balance < totalAmount {
			return nil, nil, apperror.New(http.StatusUnauthorized, "UNAUTHORIZED", "Insufficient wallet balance.")
		}
		orderData.Status = order.OrderStatusConfirmed
	default:
		orderData.Status = order.OrderStatusPending
	}

	// 14. create order && update stock
	if err := s.orderRepo.CreateOrder(ctx, orderData); err != nil {
		s.log.Error("Failed to create order", zap.Error(err))
		if err == apperror.ErrNotEnoughStock {
			return nil, nil, apperror.New(http.StatusConflict, "NOT_ENOUGH_STOCK", "Not enough stock.")
		}
		return nil, nil, apperror.New(http.StatusInternalServerError, "DB_ERROR", "Failed to create order.")
	}

	var _ payment.PaymentResponse
	// 15. create payment and payment response
	paymentResp, err := gateway.CreatePayment(ctx, payment.PaymentRequest{
		OrderID:  orderID,
		UserID:   userID,
		Amount:   int64(totalAmount) / 100,
		Currency: "INR",
	})
	if err != nil {
		s.log.Error("Failed to initialize payment", zap.Error(err))
		return nil, nil, apperror.New(http.StatusInternalServerError, "DB_ERROR", "Failed to initialize payment.")
	}

	if paymentResp.GatewayRef != nil {
		fmt.Println("Payment created : ", *paymentResp.GatewayRef)
	}

	if paymentResp.PaymentURL != nil {
		fmt.Println("Payment URL : ", *paymentResp.PaymentURL)
	}

	paymentObj := &payment.Payment{
		OrderID:  orderID,
		UserID:   userID,
		Amount:   int64(totalAmount) * 100,
		Currency: "INR",
		Provider: req.PaymentMethod,
		Status:   paymentResp.Status,
	}

	if err := s.paymentRepo.CreatePayment(ctx, paymentObj); err != nil {
		s.log.Error("Failed to create payment", zap.Error(err))
		return nil, nil, apperror.New(http.StatusInternalServerError, "DB_ERROR", "Failed to create payment.")
	}

	neworder, err := s.orderRepo.GetUserOrderByID(ctx, orderID, 0)
	if err != nil {
		s.log.Error("Failed to get order", zap.Error(err))
		if err == apperror.ErrOrderNotFound {
			return nil, nil, apperror.New(http.StatusNotFound, "NOT_FOUND", apperror.ErrOrderNotFound.Error())
		}
		return nil, nil, apperror.New(http.StatusInternalServerError, "DB_ERROR", "Failed to get order.")
	}

	// return response
	resp := &order.CreateOrderResponse{
		OrderID:               orderID,
		Items:                 neworder.Items,
		Subtotal:              cartObj.CartTotalPrice,
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
		ShipmentStatus:        string(shipping.ShipmentStatusPending),
		PaymentMethod:         string(req.PaymentMethod),
		PaymentStatus:         string(paymentResp.Status),
		Payment:               paymentObj,
		CreatedAt:             time.Now(),
	}

	resp.CouponData = cartObj.CouponData

	if req.PaymentMethod == payment.PaymentMethodCOD {
		resp.Payment = nil
	}
	s.log.Info("Order created successfully", zap.Any("order", resp))

	// clear cart if order creation is successful
	if req.PaymentMethod == payment.PaymentMethodCOD {
		s.cartRepo.EmptyCart(ctx, userID)
	}

	return resp, nil, nil
}

func (s *OrderService) UpdateOrderStatusOnPayment(ctx context.Context, status payment.PaymentStatus, orderID string) *apperror.APIError {

	// 1. validate payment status
	if status != payment.PaymentStatusCompleted && status != payment.PaymentStatusFailed {
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
func (s *OrderService) GetUserOrders(ctx context.Context) ([]order.OrderBaseResponse, *apperror.APIError) {
	utils.LogCtxContent(ctx, s.log)
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

	fmt.Printf("Orders Amount: %f\n", orders[0].TotalAmount)

	return orders, nil
}

func (s *OrderService) GetOrderByID(ctx context.Context, orderID string) (*order.OrderResponse, *apperror.APIError) {

	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		s.log.Error("Failed to get user ID", zap.Error(err))
		return nil, &apperror.APIError{
			Status:  http.StatusUnauthorized,
			Code:    "UNAUTHORIZED",
			Message: "You are not authorized to perform this action.",
		}
	}

	orderObj, err := s.orderRepo.GetUserOrderByID(ctx, orderID, userID)
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

	// deductableAmount := 0.0
	// for _, v := range order.Items {
	// 	if v.Status == "cancelled" {
	// 		deductableAmount += v.TotalPrice
	// 	}
	// }

	// order.ShippingCost = domain.DeliveryTypeCharges[order.DeliveryType]
	// order.TotalAmount = order.Subtotal + order.TaxAmount + order.ShippingCost
	// order.PayableAmount = order.Subtotal + order.TaxAmount + order.ShippingCost - deductableAmount

	orderObj = utils.CalculateOrderTotalAmount(orderObj)

	return orderObj, nil
}

// cancel order
func (s *OrderService) CancelOrder(ctx context.Context, req *order.CancelOrderRequest) (*order.OrderResponse, *apperror.APIError) {

	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, &apperror.APIError{
			Status:  http.StatusUnauthorized,
			Code:    "UNAUTHORIZED",
			Message: "You are not authorized to perform this action.",
		}
	}

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

	var _ order.OrderResponse
	orderObj, err := s.orderRepo.GetUserOrderByID(ctx, req.OrderID, userID)
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

	statusStr := strings.ToUpper(string(orderObj.Status))

	switch statusStr {
	case string(order.OrderStatusCancelled):
		return nil, &apperror.APIError{
			Status:  http.StatusConflict,
			Code:    "NOT_VALID",
			Message: "Order is already cancelled.",
		}
	case string(order.OrderStatusDelivered):
		return nil, &apperror.APIError{
			Status:  http.StatusConflict,
			Code:    "NOT_VALID",
			Message: "Order is already delivered.Choose to return the order.",
		}
	case string(order.OrderStatusPending):
		return nil, &apperror.APIError{
			Status:  http.StatusConflict,
			Code:    "NOT_VALID",
			Message: "Order has not been confirmed yet.",
		}
	}

	err = s.orderRepo.CancelOrderNew(ctx, req.OrderID, req.Reason, req.OrderItemsID)
	if err != nil {
		s.log.Error("Failed to cancel order", zap.Error(err))
		return nil, &apperror.APIError{
			Status:  http.StatusConflict,
			Code:    "DB_ERROR",
			Message: "Failed to cancel order.",
		}
	}

	s.log.Info("Order cancelled successfully", zap.String("order_id", req.OrderID))

	if orderObj.PaymentMethod == string(payment.PaymentMethodRazorpay) {
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

	updateOrder, err := s.orderRepo.GetUserOrderByID(ctx, req.OrderID, userID)
	if err != nil {
		s.log.Error("Failed to get order", zap.Error(err))
		return nil, &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "DB_ERROR",
			Message: "Failed to get order.",
		}
	}

	var deductableAmount float64
	for _, v := range updateOrder.Items {
		if v.Status == "cancelled" {
			deductableAmount += v.TotalPrice
		}
	}

	updateOrder.ShippingCost = shipping.DeliveryTypeCharges[updateOrder.DeliveryType]
	updateOrder.TotalAmount = updateOrder.Subtotal + updateOrder.TaxAmount + updateOrder.ShippingCost
	updateOrder.PayableAmount = updateOrder.Subtotal + updateOrder.TaxAmount + updateOrder.ShippingCost - deductableAmount

	return updateOrder, nil
}

// ORDER ADMIN SERVICES
// ///////////////////////
// list all orders
func (s *OrderService) ListAllOrders(ctx context.Context, filter *order.OrderFilter) ([]order.OrderBaseResponse, *apperror.APIError) {

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

func (s *OrderService) ListReturnRequests(ctx context.Context, filter *order.ReturnFilter) ([]order.BaseReturnResponse, *apperror.APIError) {

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
	if *filter.Sort != "ASC" && *filter.Sort != "DESC" {
		return nil, &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_SORT",
			Message: "Invalid sort.",
		}
	}

	var validOrderBy = []string{"refunded_amount", "date"}

	if filter.OrderBy != nil && !utils.IsValueValid(*filter.OrderBy, validOrderBy) {
		return nil, &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_ORDER_BY",
			Message: "Invalid order by.",
		}
	}

	if *filter.OrderBy == "date" {
		*filter.OrderBy = "ors.created_at"
	} else {
		*filter.OrderBy = "ors.refunded_amount"
	}

	if filter.Status != nil && !utils.IsValueValid(*filter.Status, order.ValidReturnStatus) {
		return nil, &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_STATUS",
			Message: "Invalid status.",
		}
	}

	returns, err := s.orderRepo.ListAllReturns(ctx, filter)
	if err != nil {
		s.log.Error("Failed to get returns", zap.Error(err))
		return nil, &apperror.APIError{
			Status:  http.StatusConflict,
			Code:    "DB_ERROR",
			Message: "Failed to get returns.",
		}
	}
	return returns, nil
}

// update order status
func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderID string, status shipping.ShipmentStatus) (*order.OrderResponse, *apperror.APIError) {

	// check if order exists
	if status != "shipped" && status != "delivered" {
		return nil, &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_STATUS",
			Message: "Invalid status.",
		}
	}

	orderObj, err := s.orderRepo.GetUserOrderByID(ctx, orderID, 0)
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

	strOrdStatus := strings.ToUpper(string(orderObj.Status)) //
	strShipmentStatus := orderObj.ShipmentStatus

	// check if status is valid for the order
	switch strOrdStatus {
	case string(order.OrderStatusCancelled):
		return nil, &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_STATUS",
			Message: "Order is already cancelled.",
		}
	case string(order.OrderStatusDelivered):
		return nil, &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_STATUS",
			Message: "Order is already delivered.",
		}
	case string(order.OrderStatusPending):
		return nil, &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_STATUS",
			Message: "Order has not been confirmed.",
		}
	case string(order.OrderStatusFailed):
		return nil, &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_STATUS",
			Message: "Order has not been confirmed.",
		}
	}

	fmt.Println("comparing statuse", strOrdStatus)
	switch strOrdStatus {
	case string(shipping.ShipmentStatusShipped):
		if orderObj.ShipmentStatus == strings.ToLower(string(shipping.ShipmentStatusShipped)) && status == "shipped" {
			return nil, &apperror.APIError{
				Status:  http.StatusNotFound,
				Code:    "INVALID_STATUS",
				Message: "Order is already shipped.",
			}
		}
	case string(shipping.ShipmentStatusDelivered):
		if orderObj.ShipmentStatus == strings.ToLower(string(shipping.ShipmentStatusDelivered)) {
			return nil, &apperror.APIError{
				Status:  http.StatusNotFound,
				Code:    "INVALID_STATUS",
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
	if orderObj.PaymentMethod == "COD" {
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

// returns

func (s *OrderService) ReturnOrderItemRequest(ctx context.Context, req *order.ReturnOrderItemRequest) *apperror.APIError {

	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return &apperror.APIError{
			Status:  http.StatusUnauthorized,
			Code:    "UNAUTHORIZED",
			Message: "You are not authorized to perform this action.",
		}
	}

	if req.OrderID == "" || req.ItemID <= 0 {
		return &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "BAD_REQUEST",
			Message: "Invalid request.",
		}
	}

	if req.Reason == "" || len(req.Reason) < 10 {
		return &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "BAD_REQUEST",
			Message: "Please provide a valid return reason.",
		}
	}

	orderObj, err := s.orderRepo.GetUserOrderByID(ctx, req.OrderID, userID)
	if err != nil {
		s.log.Error("Failed to get order", zap.Error(err))
		if err == apperror.ErrOrderNotFound {
			return &apperror.APIError{
				Status:  http.StatusNotFound,
				Code:    "NOT_FOUND",
				Message: apperror.ErrOrderNotFound.Error(),
			}
		}
		return &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "DB_ERROR",
			Message: "Failed to get order.",
		}
	}

	strStatus := strings.ToUpper(string(orderObj.Status))

	if strStatus != string(order.OrderStatusDelivered) {
		return &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "BAD_REQUEST",
			Message: "Order can be only returned after delivery.",
		}
	}

	for _, item := range orderObj.Items {
		if item.ItemID == req.ItemID {
			if item.Status != "delivered" {
				return &apperror.APIError{
					Status:  http.StatusBadRequest,
					Code:    "BAD_REQUEST",
					Message: "Order item can be only returned after delivery.",
				}
			}
		}
	}

	if err := s.orderRepo.ReturnOrderItemRequest(ctx, req.OrderID, userID, req.ItemID, req.Reason); err != nil {
		s.log.Error("Failed to return order item", zap.Error(err))
		if err == apperror.ErrReturnItemNotFound {
			return &apperror.APIError{
				Status:  http.StatusNotFound,
				Code:    "NOT_FOUND",
				Message: apperror.ErrReturnItemNotFound.Error(),
			}
		}
		return &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "DB_ERROR",
			Message: "Failed to return order item.",
		}
	}

	return nil
}

func (s *OrderService) ReturnOrderRequest(ctx context.Context, req *order.ReturnOrderRequest) *apperror.APIError {

	// Getting userID from context
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return &apperror.APIError{
			Status:  http.StatusUnauthorized,
			Code:    "UNAUTHORIZED",
			Message: "You are not authorized to perform this action.",
		}
	}

	// Basic request validation
	if req.OrderID == "" {
		return &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "BAD_REQUEST",
			Message: "Invalid request.",
		}
	}

	// Reason validation
	if req.Reason == "" || len(req.Reason) < 5 {
		return &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "BAD_REQUEST",
			Message: "Please provide a valid return reason.",
		}
	}

	// Fetching order with user
	orderObj, err := s.orderRepo.GetUserOrderByID(ctx, req.OrderID, userID)
	if err != nil {
		s.log.Error("Failed to get order", zap.Error(err))
		if err == apperror.ErrOrderNotFound {
			return &apperror.APIError{
				Status:  http.StatusNotFound,
				Code:    "NOT_FOUND",
				Message: apperror.ErrOrderNotFound.Error(),
			}
		}
		return &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "DB_ERROR",
			Message: "Failed to get order.",
		}
	}

	// Converting status to string
	strStatus := strings.ToUpper(string(orderObj.Status))

	if strStatus != string(order.OrderStatusDelivered) {
		return &apperror.APIError{
			Status:  http.StatusConflict,
			Code:    "INVALID_STATUS",
			Message: "Order can be only returned after delivery.",
		}
	}

	if err := s.orderRepo.ReturnOrderRequest(ctx, req.OrderID, userID, req.Reason); err != nil {
		s.log.Error("Failed to create return request", zap.Error(err))
		return &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "DB_ERROR",
			Message: "Failed to create return request.",
		}
	}

	return nil
}

func (s *OrderService) GetReturnRequest(ctx context.Context, returnID int64) (*order.FullReturnResponse, *apperror.APIError) {

	res, err := s.orderRepo.GetReturnRequest(ctx, returnID)
	if err != nil {
		s.log.Error("Failed to get return request", zap.Error(err))
		if err == apperror.ErrReturnRequestNotFound {
			return nil, &apperror.APIError{
				Status:  http.StatusNotFound,
				Code:    "NOT_FOUND",
				Message: apperror.ErrReturnRequestNotFound.Error(),
			}
		}
		return nil, &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "DB_ERROR",
			Message: "Failed to get return request.",
		}
	}

	return res, nil
}

func (s *OrderService) UpdateReturnRequestStatus(ctx context.Context, req *order.UpdateReturnRefundRequest) (*order.FullReturnResponse, *apperror.APIError) {

	if req.Status != "approved" && req.Status != "rejected" {
		return nil, &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_STATUS",
			Message: "Please provide a valid status.",
		}
	}

	// Fetching return request
	returnRequest, err := s.orderRepo.GetReturnRequest(ctx, req.ReturnID)
	if err != nil {
		return nil, &apperror.APIError{
			Status:  http.StatusNotFound,
			Code:    "NOT_FOUND",
			Message: apperror.ErrReturnRequestNotFound.Error(),
		}
	}

	// Current return status check
	returnStatus := returnRequest.Status

	switch returnStatus {

	case "returned":
		return nil, &apperror.APIError{
			Status:  http.StatusConflict,
			Code:    "RETURN_ALREADY_RETURNED",
			Message: "Return has been already returned",
		}

	case "rejected":
		return nil, &apperror.APIError{
			Status:  http.StatusConflict,
			Code:    "RETURN_ALREADY_REJECTED",
			Message: "Return has been already rejected",
		}
	}

	// Updating return request status
	if err := s.orderRepo.UpdateReturnRequestStatus(ctx, req); err != nil {
		s.log.Error("Failed to update return request status", zap.Error(err))
		if err == apperror.ErrInvalidReturnState {
			return nil, &apperror.APIError{
				Status:  http.StatusConflict,
				Code:    "INVALID_STATUS",
				Message: "Return has not been returned yet",
			}
		}
		return nil, &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "DB_ERROR",
			Message: "Failed to update return request status.",
		}
	}

	returnRequest, err = s.orderRepo.GetReturnRequest(ctx, req.ReturnID)
	if err != nil {

		s.log.Error("Failed to get return request", zap.Error(err))
		return nil, &apperror.APIError{
			Status:  http.StatusNotFound,
			Code:    "NOT_FOUND",
			Message: apperror.ErrReturnRequestNotFound.Error(),
		}
	}

	return returnRequest, nil
}

func (s *OrderService) ProcessReturnRefundRequest(ctx context.Context, req *order.UpdateReturnRefundRequest) (*order.FullReturnResponse, *apperror.APIError) {

	// Fetching the return request
	returnRequest, err := s.orderRepo.GetReturnRequest(ctx, req.ReturnID)
	if err != nil {
		return nil, &apperror.APIError{
			Status:  http.StatusNotFound,
			Code:    "NOT_FOUND",
			Message: apperror.ErrReturnRequestNotFound.Error(),
		}
	}

	switch returnRequest.Status {

	case "rejected":
		return nil, &apperror.APIError{
			Status:  http.StatusConflict,
			Code:    "INVALID_STATUS",
			Message: "Return has been already rejected",
		}

	case "requested":
		return nil, &apperror.APIError{
			Status:  http.StatusConflict,
			Code:    "INVALID_STATUS",
			Message: "Return has not been verified yet",
		}

	case "refunded":
		return nil, &apperror.APIError{
			Status:  http.StatusConflict,
			Code:    "INVALID_STATUS",
			Message: "Return has been already returned",
		}
	}

	req.RefundAmount = returnRequest.TotalRefundValue
	if err := s.orderRepo.ProcessReturnRefund(ctx, req); err != nil {
		s.log.Error("Failed to process return refund request", zap.Error(err))
		return nil, &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "DB_ERROR",
			Message: "Failed to process return refund request.",
		}
	}

	returnRequest, err = s.orderRepo.GetReturnRequest(ctx, req.ReturnID)
	if err != nil {
		s.log.Error("Failed to get return request", zap.Error(err))
		return nil, &apperror.APIError{
			Status:  http.StatusNotFound,
			Code:    "NOT_FOUND",
			Message: apperror.ErrReturnRequestNotFound.Error(),
		}
	}

	return returnRequest, nil
}
