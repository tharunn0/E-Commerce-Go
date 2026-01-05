package domain

import (
	"context"
	"time"
)

type UserRepository interface {
	// auth operations
	RegisterUser(ctx context.Context, req *RegisterRequest) error
	GetUser(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, userID int64) (*User, error)
	GoogleSignIn(ctx context.Context, req *GoogleSignInRequest) (*User, error)

	// profile operations
	UpdateUserProfile(ctx context.Context, userID int64, req *UpdateUserProfileRequest) (*UserProfile, error)

	// address operations
	InsertUserAddress(ctx context.Context, address *UserAddress) error
	GetUserAddresses(ctx context.Context, userID int64) ([]*UserAddress, error)
	GetUserAddressByID(ctx context.Context, addressID int64) (*UserAddress, error)
	GetDefaultUserAddress(ctx context.Context, userID int64) (*UserAddress, error)
	UpdateDefaultUserAddress(ctx context.Context, userID int64, addressID int64) error
	UpdateUserAddress(ctx context.Context, userID int64, req *UpdateUserAddressRequest) (*UserAddress, error)
	DeleteUserAddress(ctx context.Context, userID int64, addressID int64) error
}

type AdminRepository interface {
	CreateAdmin(ctx context.Context, req *AdminRegisterRequest) error
	GetAdmin(ctx context.Context, email string) (*User, error)

	// GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUsersByID(ctx context.Context, id int64) (*UserProfile, error)
	ListUsers(ctx context.Context, filter *UserFilter) ([]*UserProfile, error)
	UpdateUserStatus(ctx context.Context, req *UserStatusUpdateRequest) error
	DeleteUser(ctx context.Context, id int64) error
}

type AuthTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type User struct {
	ID               int64      `json:"id"`
	Email            string     `json:"email"`
	Phone            string     `json:"phone"`
	Password         string     `json:"password"`
	FirstName        string     `json:"firstName"`
	LastName         string     `json:"lastName"`
	Role             string     `json:"role,omitempty"`
	IsVerified       bool       `json:"isVerified"`
	Status           string     `json:"status"`
	DefaultAddressID *int64     `json:"defaultAddressId,omitempty"`
	ProfilePicture   *string    `json:"profilePicture,omitempty"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
	DeletedAt        *time.Time `json:"deletedAt,omitempty"`
}
type UserProfile struct {
	ID               int64      `json:"id"`
	Email            string     `json:"email"`
	FirstName        string     `json:"firstName"`
	LastName         string     `json:"lastName"`
	Phone            *string    `json:"phone"`
	Role             *string    `json:"role,omitempty"`
	IsVerified       bool       `json:"isVerified"`
	Status           *string    `json:"status,omitempty"`
	DefaultAddressID *int64     `json:"defaultAddressId,omitempty"`
	ProfilePicture   *string    `json:"profilePicture,omitempty"`
	CreatedAt        *time.Time `json:"createdAt"`
	UpdatedAt        *time.Time `json:"updatedAt,omitempty"`

	Addresses []*UserAddress `json:"addresses,omitempty"`
}
type UserAddress struct {
	ID           int64      `json:"id"`
	UserID       int64      `json:"user_id,omitempty"`
	Label        string     `json:"label,omitempty"`
	AddressLine  string     `json:"address_line"`
	AddressLine2 string     `json:"address_line_2,omitempty"`
	Pincode      string     `json:"pincode"`
	City         string     `json:"city"`
	District     string     `json:"district"`
	State        string     `json:"state"`
	Country      string     `json:"country"`
	CreatedAt    *time.Time `json:"created_at,omitempty"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty"`
}

type UpdateUserAddressRequest struct {
	ID           int64   `json:"id"`
	Label        *string `json:"label,omitempty"`
	AddressLine  *string `json:"address_line,omitempty"`
	AddressLine2 *string `json:"address_line_2,omitempty"`
	Pincode      *string `json:"pincode,omitempty"`
	City         *string `json:"city,omitempty"`
	District     *string `json:"district,omitempty"`
	State        *string `json:"state,omitempty"`
	Country      *string `json:"country,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}
type RegisterRequest struct {
	Email           string `json:"email" validate:"required,email"`
	Phone           string `json:"phone" validate:"omitempty"`
	Password        string `json:"password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required,min=8"`
	FirstName       string `json:"firstName" validate:"required"`
	LastName        string `json:"lastName" validate:"required"`
}

type AdminRegisterRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Phone     string `json:"phone" validate:"omitempty"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
	AdminCode string `json:"admin_code"`
}

type GoogleSignInRequest struct {
	Email     string
	Token     string
	Verified  bool
	FirstName string
	LastName  string
	Sub       string
}

type LoginResponse struct {
	User struct {
		ID        int64  `json:"id"`
		Email     string `json:"email"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		Role      string `json:"role"`
	} `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type UserFilter struct {
	Status string `json:"status,omitempty"`
	Role   string `json:"role,omitempty"`
	Search string `json:"search,omitempty"`
	Limit  int    `json:"limit,omitempty"`
	Page   int    `json:"page,omitempty"`
	Total  int    `json:"total,omitempty`
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

type UserStatusUpdateRequest struct {
	UserID int64  `json:"user_id" validate:"required"`
	Status string `json:"status" validate:"required,oneof=active inactive"`
}
type UpdateUserProfileRequest struct {
	FirstName      *string `json:"first_name,omitempty"`
	LastName       *string `json:"last_name,omitempty"`
	Phone          *string `json:"phone,omitempty"`
	ProfilePicture *string `json:"profile_img_url,omitempty"`
}
