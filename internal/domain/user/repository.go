package user

import (
	"context"

	"github.com/tharunn0/E-Commerce-Go/internal/domain/payment"
)

type UserRepository interface {
	// auth operations
	RegisterUser(ctx context.Context, req *RegisterRequest, refCode string) error
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

	GetWallet(ctx context.Context, userID int64) (*payment.Wallet, error)
	GetWalletTransactions(ctx context.Context, userID int64, filter *payment.TransactionFilter) ([]*payment.WalletTransaction, error)
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
