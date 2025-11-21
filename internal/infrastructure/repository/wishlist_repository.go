package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tharunn0/E-Commerce-Go/internal/apperror"
	"github.com/tharunn0/E-Commerce-Go/internal/domain"
)

type WishlistRepository struct {
	DB *pgxpool.Pool
}

func NewWishlistRepository(db *pgxpool.Pool) *WishlistRepository {
	return &WishlistRepository{
		DB: db,
	}
}

func uqConstraintViolation(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" { // unique violation
			return pgErr.ConstraintName
		}
	}
	return ""
}

func (repo *WishlistRepository) AddToWishlist(ctx context.Context, userID int64, productID int64) (int64, error) {

	var wishlistId int64
	query := `SELECT id FROM wishlists WHERE user_id = $1`
	err := repo.DB.QueryRow(ctx, query, userID).Scan(&wishlistId)
	if err != nil {
		if err == pgx.ErrNoRows {
			err := repo.DB.QueryRow(ctx, `INSERT INTO wishlists (user_id) VALUES ($1) RETURNING id`, userID).Scan(&wishlistId)
			if err != nil {
				return 0, apperror.ErrWishlistCreateFail
			}
		} else {
			return 0, err
		}
	}

	query = `INSERT into wishlist_items (wishlist_id, product_id) VALUES ($1, $2) RETURNING id`
	_, err = repo.DB.Exec(ctx, query, wishlistId, productID)
	if err != nil {
		if msg := uqConstraintViolation(err); msg == "uq_wishlist_id_product_id" {
			return 0, apperror.ErrWishlistItemAlreadyExists
		}
		return 0, err
	}

	return wishlistId, err
}

func (repo *WishlistRepository) GetWishlistByUserID(ctx context.Context, userID int64) (*domain.Wishlist, error) {

	wishlist := domain.Wishlist{}
	query := `SELECT p.id, p.name, p.min_price, p.max_price, p.description, p.image_url FROM products p 
	JOIN wishlist_items wi ON p.id = wi.product_id WHERE wi.wishlist_id = (SELECT id FROM wishlists WHERE user_id = $1)`
	rows, err := repo.DB.Query(ctx, query, userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperror.ErrWishlistNotFound
		}
		return nil, err
	}

	for rows.Next() {
		var product domain.ProductResponse
		if err := rows.Scan(&product.ID, &product.Name, &product.MinPrice, &product.MaxPrice, &product.Description, &product.ImageURL); err != nil {
			return nil, err
		}
		wishlist.Products = append(wishlist.Products, &product)
	}

	return &wishlist, nil
}

func (repo *WishlistRepository) GetWishlistByID(ctx context.Context, wishlistID int64) (*domain.Wishlist, error) {

	wishlist := domain.Wishlist{}
	query := `SELECT p.id, p.name, p.min_price, p.max_price, p.description, p.image_url FROM products p 
	JOIN wishlist_items wi ON p.id = wi.product_id WHERE wi.wishlist_id = $1`
	rows, err := repo.DB.Query(ctx, query, wishlistID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperror.ErrWishlistNotFound
		}
		return nil, err
	}

	for rows.Next() {
		var product domain.ProductResponse
		if err := rows.Scan(&product.ID, &product.Name, &product.MinPrice, &product.MaxPrice, &product.Description, &product.ImageURL); err != nil {
			return nil, err
		}
		wishlist.Products = append(wishlist.Products, &product)
	}

	return &wishlist, nil
}

func (repo *WishlistRepository) RemoveFromWishlist(ctx context.Context, userID int64, productIDs []int64) error {

	query := `DELETE FROM wishlist_items 
	WHERE wishlist_id = (SELECT id FROM wishlists WHERE user_id = $1) AND
	product_id = ANY($2)`
	cmdTag, err := repo.DB.Exec(ctx, query, userID, productIDs)
	if err != nil {
		return fmt.Errorf("failed to execute delete query: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return apperror.ErrWishlistItemNotFound
	}

	return nil
}
