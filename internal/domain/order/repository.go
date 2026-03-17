package order

import (
	"context"

	"github.com/tharunn0/E-Commerce-Go/internal/domain/payment"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/shipping"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, data *CreateOrderData) error
	GetUserOrders(ctx context.Context, userID int64) ([]OrderBaseResponse, error)
	GetUserOrderByID(ctx context.Context, orderID string, userID int64) (*OrderResponse, error)

	// cancellation
	CancelOrder(ctx context.Context, orderID string, reason string, orderItemsID []int64) error
	CancelOrderNew(ctx context.Context, orderID string, reason string, orderItemsID []int64) error

	// returns
	ReturnOrderItemRequest(ctx context.Context, orderID string, userID int64, itemID int64, reason string) error
	ReturnOrderRequest(ctx context.Context, orderID string, userID int64, reason string) error

	// updations
	UpdateOrderStatusOnPayment(ctx context.Context, orderID string, status payment.PaymentStatus) error
	UpdateShipmentStatus(ctx context.Context, orderID string, status shipping.ShipmentStatus, cod bool) error

	// Admin ops
	UpdateReturnRequestStatus(ctx context.Context, req *UpdateReturnRefundRequest) error
	ProcessReturnRefund(ctx context.Context, req *UpdateReturnRefundRequest) error
	// shipment
	ListAllOrders(ctx context.Context, filter *OrderFilter) ([]OrderBaseResponse, error)      //admin
	ShipOrder(ctx context.Context, orderID string, shipmentData *shipping.ShipmentData) error //admin
	DeliverOrder(ctx context.Context, orderID string) error                                   //admin

	ListAllReturns(ctx context.Context, filter *ReturnFilter) ([]BaseReturnResponse, error) //admin
	GetReturnRequest(ctx context.Context, returnID int64) (*FullReturnResponse, error)      //admin
}
