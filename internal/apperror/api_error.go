package apperror

import "errors"

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Generic errors
var (
	ErrInvalidInput     = errors.New("Invalid input")
	ErrUnauthorized     = errors.New("Unauthorized")
	ErrForbidden        = errors.New("Forbidden")
	ErrNotFound         = errors.New("Not found")
	ErrConflict         = errors.New("Conflict")
	ErrInternal         = errors.New("Internal server error")
	ErrDatabase         = errors.New("Database error")
	ErrValidationFailed = errors.New("Validation failed")
)

// User-related errors
var (
	ErrEmailExists     = errors.New("Email already exists")
	ErrPhoneExists     = errors.New("Phone number already exists")
	ErrUserCreateFail  = errors.New("User creation failed")
	ErrUserNotFound    = errors.New("User not found")
	ErrInvalidPassword = errors.New("Invalid password")
)

// Address-related errors
var (
	ErrAddressNotFoundForUser = errors.New("Address not found for user")
)

// Auth / Token errors
var (
	ErrTokenExpired   = errors.New("Token expired")
	ErrTokenInvalid   = errors.New("Invalid token")
	ErrSessionExpired = errors.New("Session expired")
	ErrOTPExpired     = errors.New("OTP has expired")
)

// Business rule errors (optional)
var (
	ErrInsufficientBalance = errors.New("Insufficient balance")
	ErrAlreadyProcessed    = errors.New("Already processed")
	ErrOperationNotAllowed = errors.New("Operation not allowed")
)
