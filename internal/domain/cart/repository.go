package cart

import (
	"context"
)

//go:generate mockgen -destination=mocks/mock_cart_repo.go -package=mocks . CartRepository
type CartRepository interface {
	AddToCart(ctx context.Context, userID int64, productVariantID int64, quantity int64) (*int64, error)
	GetCartByID(ctx context.Context, cartID int64) (*Cart, error)
	GetCartByUserID(ctx context.Context, userID int64) (*Cart, error)
	RemoveCartItem(ctx context.Context, userID int64, productVariantID int64) (int64, error)
	UpdateCartItemQuantity(ctx context.Context, userID int64, req *UpdateCartItemQuantityRequest) (int64, error)
	EmptyCart(ctx context.Context, userID int64) error
	GetCartVariantInfo(ctx context.Context, userID int64) (map[int64]VariantInfo, error)
}
