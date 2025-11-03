package domain

import (
	"context"
	"time"
)

type UserRepository interface {
	RegisterUser(ctx context.Context, req *RegisterRequest) error
	GetUser(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, userID int64) (*User, error)
	InsertUserAddress(ctx context.Context, address *UserAddress) error
	GetUserAddresses(ctx context.Context, userID int64) ([]*UserAddress, error)
}

type AdminRepository interface {
	GetAdmin(ctx context.Context, email string) (*User, error)

	// GetUserByEmail(ctx context.Context, email string) (*User, error)
	// GetUsersByName(ctx context.Context, name string) ([]*User, error)

	ListUsers(ctx context.Context, filter *UserFilter) ([]*User, error)
	// CountUsers(ctx context.Context, filter *UserFilter) (int64, error)

	UpdateUserStatus(ctx context.Context, req *UserStatusUpdateRequest) error
}

type User struct {
	ID               int64      `json:"id"`
	Email            string     `json:"email"`
	Phone            string     `json:"phone"`
	Password         string     `json:"password"`
	FirstName        string     `json:"firstName"`
	LastName         string     `json:"lastName"`
	Role             string     `json:"role"`
	IsVerified       bool       `json:"isVerified"`
	Status           string     `json:"status"`
	DefaultAddressID *int64     `json:"defaultAddressId,omitempty"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
	DeletedAt        *time.Time `json:"deletedAt,omitempty"`
}
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}
type RegisterRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Phone     string `json:"phone" validate:"omitempty"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"firstName" validate:"required"`
	LastName  string `json:"lastName" validate:"required"`
}

type GoogleSignInRequest struct {
	Token string `json:"token" validate:"required"`
}

type LoginResponse struct {
	User struct {
		ID        int64  `json:"id"`
		Email     string `json:"email"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		Role      string `json:"role"`
	} `json:"user"`
	Token string `json:"token"`
}

type UserFilter struct {
	Status string `json:"status,omitempty"`
	Role   string `json:"role,omitempty"`
	Search string `json:"search,omitempty"`
	Limit  int    `json:"limit,omitempty"`
	Page   int    `json:"page,omitempty"`
}

type PasswordResetRequest struct {
	Email string `json:"email" validate:"required,email"`
}
type PasswordResetData struct {
	Email       string `json:"email" validate:"required,email"`
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"newPassword" validate:"required,min=8"`
}
type PasswordResetToken struct {
	Token    string    `json:"token" validate:"required"`
	ExpiryAt time.Time `json:"expiryAt" validate:"required"`
}

type UserProfile struct {
	ID               int64     `json:"id"`
	Email            string    `json:"email"`
	FirstName        string    `json:"firstName"`
	LastName         string    `json:"lastName"`
	IsVerified       bool      `json:"isVerified"`
	Status           string    `json:"status"`
	DefaultAddressID *int64    `json:"defaultAddressId,omitempty"`
	CreatedAt        time.Time `json:"createdAt"`

	Addresses []*UserAddress `json:"addresses"`
}
type UserAddress struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	Label        string    `json:"label,omitempty"`
	AddressLine  string    `json:"address_line"`
	AddressLine2 string    `json:"address_line_2,omitempty"`
	Pincode      string    `json:"pincode"`
	City         string    `json:"city"`
	State        string    `json:"state,omitempty"`
	Country      string    `json:"country"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UserStatusUpdateRequest struct {
	UserID int64  `json:"user_id" validate:"required"`
	Status string `json:"status" validate:"required,oneof=active inactive"`
}
