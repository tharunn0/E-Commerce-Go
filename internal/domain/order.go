package domain

import (
	"context"
	"time"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, data *CreateOrderData) error
	GetUserOrders(ctx context.Context, userID int64) ([]OrderBaseResponse, error)
	GetUserOrderByID(ctx context.Context, orderID string, userID int64) (*OrderResponse, error)

	// cancellation
	CancelOrder(ctx context.Context, orderID string, reason string, orderItemsID []int64) error

	// returns
	ReturnOrderItemRequest(ctx context.Context, orderID string, userID int64, itemID int64, reason string) error
	ReturnOrderRequest(ctx context.Context, orderID string, userID int64, reason string) error

	// updations
	UpdateOrderStatusOnPayment(ctx context.Context, orderID string, status PaymentStatus) error
	UpdateShipmentStatus(ctx context.Context, orderID string, status ShipmentStatus, cod bool) error

	// Admin ops
	UpdateReturnRequestStatus(ctx context.Context, req *UpdateReturnRefundRequest) error
	ProcessReturnRefund(ctx context.Context, req *UpdateReturnRefundRequest) error
	// shipment
	ListAllOrders(ctx context.Context, filter *OrderFilter) ([]OrderBaseResponse, error) //admin
	ShipOrder(ctx context.Context, orderID string, shipmentData *ShipmentData) error     //admin
	DeliverOrder(ctx context.Context, orderID string) error                              //admin

	ListAllReturns(ctx context.Context, filter *ReturnFilter) ([]BaseReturnResponse, error) //admin
	GetReturnRequest(ctx context.Context, returnID int64) (*FullReturnResponse, error)      //admin
}

type CreateOrderRequest struct {
	UserID         int64 `json:"user_id"`
	ProductVariant *struct {
		ProductVariantID int64 `json:"product_variant_id"`
		Quantity         int64 `json:"quantity"`
	} `json:"product_variant"`
	AddressID     int64         `json:"address_id"`
	DeliveryType  DeliveryType  `json:"delivery_type"`
	PaymentMethod PaymentMethod `json:"payment_method"`
}

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "PENDING"
	OrderStatusConfirmed OrderStatus = "CONFIRMED"
	OrderStatusShipped   OrderStatus = "SHIPPED"
	OrderStatusDelivered OrderStatus = "DELIVERED"
	OrderStatusCancelled OrderStatus = "CANCELLED"
	OrderStatusFailed    OrderStatus = "FAILED"
)

var OrderStatusFlow = map[OrderStatus][]OrderStatus{
	OrderStatusPending:   {OrderStatusConfirmed, OrderStatusCancelled},
	OrderStatusConfirmed: {OrderStatusShipped, OrderStatusCancelled},
	OrderStatusShipped:   {OrderStatusDelivered, OrderStatusCancelled},
	OrderStatusDelivered: {},
	OrderStatusCancelled: {},
}

type CreateOrderData struct {
	UserID                int64   `db:"user_id"`
	OrderID               string  `db:"public_order_id"`
	TotalAmount           float64 `db:"total_amount"`
	TaxAmount             float64 `db:"tax_amount"`
	Items                 []OrderItem
	ShippingAddressID     int64       `db:"shipping_address_id"`
	BillingAddressID      int64       `db:"billing_address_id"`
	Status                OrderStatus `db:"order_status"`
	DeliveryType          string      `db:"delivery_type"`
	EstimatedDeliveryDate time.Time   `db:"estimated_delivery_date"`

	OrderSource string
}

type OrderItem struct {
	ItemID           int64   `json:"item_id"`
	ProductVariantID int64   `json:"product_variant_id"`
	SKU              string  `json:"sku_at_purchase"`
	ProductName      string  `json:"product_name_at_purchase"`
	Quantity         int64   `json:"quantity"`
	UnitPrice        float64 `json:"unit_price"`
	TotalPrice       float64 `json:"total_price"`
	Status           string  `json:"item_status"`
	ImageURL         string  `json:"image_url"`
}

type CreateOrderResponse struct {
	OrderID      string      `json:"public_order_id"`
	Items        []OrderItem `json:"items"`
	Subtotal     float64     `json:"subtotal"`
	TaxAmount    float64     `json:"tax_amount"`
	ShippingCost float64     `json:"shipping_cost"`
	TotalAmount  float64     `json:"total_amount"`
	Currency     string      `json:"currency"`

	ShippingAddressID int64        `json:"shipping_address_id"`
	ShippingAddress   *UserAddress `json:"shipping_address,omitempty"`
	BillingAddressID  int64        `json:"billing_address_id"`

	DeliveryType          DeliveryType `json:"delivery_type"`
	EstimatedDeliveryTime string       `json:"estimated_delivery_time"`
	EstimatedDeliveryDate string       `json:"estimated_delivery_date"`

	Status       OrderStatus `json:"order_status"`  // order status
	ReturnStatus *string     `json:"return_status"` // return status

	ShipmentID     *int64 `json:"shipment_id,omitempty"`
	ShipmentStatus string `json:"shipment_status,omitempty"`

	PaymentMethod string `json:"payment_method"`
	PaymentStatus string `json:"payment_status"`

	Payment *Payment `json:"payment,omitempty"`

	CreatedAt time.Time `json:"created_at"`
}

type OrderBaseResponse struct {
	OrderID     string  `json:"public_order_id"`
	Subtotal    float64 `json:"subtotal"`
	TaxAmount   float64 `json:"tax_amount"`
	TotalAmount float64 `json:"total_amount"`
	Currency    string  `json:"currency"`

	ShippingAddressID int64        `json:"shipping_address_id"`
	ShippingAddress   *UserAddress `json:"shipping_address,omitempty"`

	DeliveryType          DeliveryType `json:"delivery_type"`
	EstimatedDeliveryDate time.Time    `json:"estimated_delivery_date"`

	Status       OrderStatus `json:"order_status"`
	ReturnStatus *string     `json:"return_status"` // return status

	ShipmentStatus string `json:"shipment_status"`
	PaymentStatus  string `json:"payment_status"`

	// ImageURL string `json:"image_url"`
}

type OrderResponse struct {
	OrderID       string      `json:"public_order_id"`
	Items         []OrderItem `json:"items"`
	Subtotal      float64     `json:"subtotal"`
	TaxAmount     float64     `json:"tax_amount"`
	ShippingCost  float64     `json:"shipping_cost"`
	TotalAmount   float64     `json:"total_amount"`
	PayableAmount float64     `json:"payable_amount"`
	Currency      string      `json:"currency"`

	ShippingAddressID int64 `json:"shipping_address_id"`
	BillingAddressID  int64 `json:"billing_address_id"`

	DeliveryType          DeliveryType `json:"delivery_type"`
	EstimatedDeliveryDate time.Time    `json:"estimated_delivery_date"`

	Status       OrderStatus `json:"status"`        // order status
	ReturnStatus *string     `json:"return_status"` // return status

	ShipmentID      *int64 `json:"shipment_id,omitempty"`
	ShipmentStatus  string `json:"shipment_status,omitempty"`
	ShipmentCarrier string `json:"shipment_carrier,omitempty"`
	TrackingNumber  string `json:"tracking_number,omitempty"`

	PaymentMethod string `json:"payment_method"`
	PaymentStatus string `json:"payment_status"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type OrderFilter struct {
	OrderStatus    *string `form:"order_status"`
	DeliveryType   *string `form:"delivery_type"`
	ShipmentStatus *string `form:"shipment_status"`

	OrderID *string `form:"order_id"`

	PriceFrom *float64 `form:"price_from"`
	PriceTo   *float64 `form:"price_to"`

	CreatedAtFrom *time.Time `form:"created_at_from"`
	CreatedAtTo   *time.Time `form:"created_at_to"`

	OrderBy *string `form:"order_by"`
	Sort    *string `form:"sort"`

	Page  int `form:"page"`
	Limit int `form:"limit"`
}

// type CancelOrderItemRequest struct {
// 	OrderItemID int64 `json:"variant_id"`
// }

type CancelOrderRequest struct {
	OrderID      string  `json:"order_id"`
	Reason       string  `json:"reason"`
	OrderItemsID []int64 `json:"items"`
}

type ReturnOrderItemRequest struct {
	OrderID string
	ItemID  int64
	Reason  string `json:"reason"`
}

type ReturnOrderRequest struct {
	OrderID string `json:"order_id"`
	Reason  string `json:"reason"`
}

type OrderStatusUpdateRequest struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
}

type OrderReturnRequest struct {
	OrderID string `json:"order_id"`
	Reason  string `json:"reason"`
}

type OrderReturnInfo struct {
	OrderID string `json:"order_id"`
	Reason  string `json:"reason"`
}
