package domain

import "context"

type WishlistRepository interface {
	AddToWishlist(ctx context.Context, userID int64, productVariantID int64) (int64, error)
	GetWishlistByUserID(ctx context.Context, userID int64) (*Wishlist, error)
	GetWishlistByID(ctx context.Context, wishlistID int64) (*Wishlist, error)
	RemoveFromWishlist(ctx context.Context, userID int64, productVariantIDs []int64) error
}

type Wishlist struct {
	Products []*ProductResponse `json:"products"`
}

type AddToWishlistRequest struct {
	ProductID int64 `json:"product_id"`
}

type RemoveFromWishlistRequest struct {
	ProductIDs []int64 `json:"product_ids"`
}
