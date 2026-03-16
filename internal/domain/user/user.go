package user

import (
	"time"
)

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
	ReferralCode     string     `json:"referralCode,omitempty"`
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
	ReferralCode     string     `json:"referralCode,omitempty"`

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
	ReferralCode    string `json:"referral_code"`
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
	Total  int    `json:"total,omitempty"`
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

func (req *UpdateUserProfileRequest) Validate() error {
	if req.FirstName != nil && *req.FirstName == "" {
		req.FirstName = nil
	}
	if req.LastName != nil && *req.LastName == "" {
		req.LastName = nil
	}
	if req.Phone != nil && *req.Phone == "" {
		req.Phone = nil
	}
	if req.ProfilePicture != nil && *req.ProfilePicture == "" {
		req.ProfilePicture = nil
	}
	return nil
}
