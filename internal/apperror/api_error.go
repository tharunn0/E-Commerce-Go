package apperror

import "errors"

type APIError struct {
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

var (
	// Generic errors
	ErrInvalidInput     = errors.New("Invalid input")
	ErrUnauthorized     = errors.New("Unauthorized")
	ErrForbidden        = errors.New("Forbidden")
	ErrNotFound         = errors.New("Not found")
	ErrConflict         = errors.New("Conflict")
	ErrInternal         = errors.New("Internal server error")
	ErrDatabase         = errors.New("Database error")
	ErrValidationFailed = errors.New("Validation failed")

	// User-related errors

	ErrEmailExists     = errors.New("Email already exists")
	ErrPhoneExists     = errors.New("Phone number already exists")
	ErrUserCreateFail  = errors.New("User creation failed")
	ErrUserNotFound    = errors.New("User not found")
	ErrInvalidPassword = errors.New("Invalid password")

	// Address-related errors

	ErrAddressNotFoundForUser = errors.New("Address not found for user")

	// Auth / Token errors

	ErrTokenExpired   = errors.New("Token expired")
	ErrTokenInvalid   = errors.New("Invalid token")
	ErrSessionExpired = errors.New("Session expired")
	ErrOTPExpired     = errors.New("OTP has expired")

	// Cart-related errors

	ErrCartNotFound           = errors.New("Cart not found")
	ErrProductVariantNotFound = errors.New("Product variant not found")
	ErrCartEmpty              = errors.New("Cart is empty")
	ErrCartItemNotFound       = errors.New("Cart item not found")
	ErrQuantityExceeded       = errors.New("Quantity exceeded. Max 5 items allowed per variant")

	// Product variant-related errors
	ErrNoStock  = errors.New("No stock available")
	ErrLowStock = errors.New("Stock is low")

	// Wishlist-related errors
	ErrWishlistNotFound          = errors.New("Wishlist not found")
	ErrWishlistCreateFail        = errors.New("Wishlist creation failed")
	ErrWishlistDeleteFail        = errors.New("Wishlist deletion failed")
	ErrWishlistUpdateFail        = errors.New("Wishlist update failed")
	ErrWishlistItemAlreadyExists = errors.New("Wishlist item already exists")
	ErrWishlistItemNotFound      = errors.New("Wishlist item not found")

	// Order-related errors
	ErrOrderNotFound       = errors.New("Order not found")
	ErrOrderCreateFail     = errors.New("Order creation failed")
	ErrOrderUpdateFail     = errors.New("Order update failed")
	ErrOrderDeleteFail     = errors.New("Order deletion failed")
	ErrOrderAmountExceeded = errors.New("Order amount exceeded. Should be less than 1000000")

	ErrOrderItemNotFound = errors.New("Order item not found")

	// Order-create related errors
	ErrCreateOrderFail = errors.New("Order creation failed")

	// Payment related errors
	ErrPaymentCreateFail = errors.New("Payment creation failed")
	ErrPaymentNotFound   = errors.New("Payment not found")

	ErrReturnRequestNotFound  = errors.New("Return request not found")
	ErrReturnItemNotFound     = errors.New("Return item not found")
	ErrReturnAlreadyProcessed = errors.New("Return has been already processed")
	ErrInvalidReturnState     = errors.New("Invalid return state")
)
