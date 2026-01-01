package domain

import "time"

var ValidReturnStatus = []string{"requested", "partially_approved", "approved", "partially_received",
	"received", "partially_refunded", "refunded", "rejected", "cancelled"}

type ReturnFilter struct {
	OrderID     *string    `form:"order_id"`
	Status      *string    `form:"status"`
	CreatedFrom *time.Time `form:"created_at_from"`
	CreatedTo   *time.Time `form:"created_at_to"`
	OrderBy     *string    `form:"order_by"`
	Sort        *string    `form:"sort"`
	Page        int        `form:"page"`
	Limit       int        `form:"limit"`
}

type BaseReturnResponse struct {
	ID             int64   `json:"id"`
	OrderID        string  `json:"order_id"`
	Status         string  `json:"status"`
	RefundedAmount float64 `json:"refunded_amount"`
	TotalItems     int     `json:"total_items"`
	// TotalQuantity    int       `json:"total_quantity"`
	TotalRefundValue float64   `json:"total_refundable_amount"`
	CreatedAt        time.Time `json:"created_at"`
}

type FullReturnResponse struct {
	ID               int64        `json:"id"`
	OrderID          string       `json:"order_id"`
	User             UserSummary  `json:"user"`
	Status           string       `json:"status"`
	RefundedAmount   float64      `json:"refunded_amount"`
	TotalItems       int          `json:"total_items"`
	TotalQuantity    int          `json:"total_quantity"`
	TotalRefundValue float64      `json:"total_refundable_amount"`
	CreatedAt        time.Time    `json:"created_at"`
	Items            []ReturnItem `json:"items"`
}

type UserSummary struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type ReturnItem struct {
	ItemOrderID    int64      `json:"order_item_id"`
	ProductName    string     `json:"product_name_at_purchase"`
	SKU            string     `json:"sku_at_purchase"`
	Quantity       int        `json:"quantity"`
	OrderItemPrice float64    `json:"order_item_price"`
	Status         string     `json:"status"`
	Reason         string     `json:"reason"`
	ImageURL       string     `json:"image_url"`
	RequestedAt    time.Time  `json:"requested_at"`
	ApprovedAt     *time.Time `json:"approved_at,omitempty"`
}

type UpdateReturnRequest struct {
	ReturnID int64  `json:"id"`
	Status   string `json:"status"`
}
type UpdateReturnRefundRequest struct {
	ReturnID     int64  `json:"id"`
	UserID       int64  `json:"user_id"`
	Status       string `json:"status"`
	Remarks      string `json:"remarks"`
	RelatedOrder int64
	RefundAmount float64
}

var ReturnStatusFlow = map[string][]string{
	"requested":          {"approved", "rejected", "cancelled"},
	"approved":           {"received", "partially_received", "partially_refunded", "rejected"},
	"received":           {"partially_refunded", "refunded", "rejected", "cancelled"},
	"partially_received": {"partially_refunded", "refunded", "rejected", "cancelled"},
	"partially_refunded": {"refunded", "rejected", "cancelled"},
	"refunded":           {"rejected", "cancelled"},
	"rejected":           {"cancelled"},
	"cancelled":          {},
}
