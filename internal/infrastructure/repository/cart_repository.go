package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tharunn0/E-Commerce-Go/internal/apperror"
	"github.com/tharunn0/E-Commerce-Go/internal/domain"
	// "github.com/tharunn0/E-Commerce-Go/internal/domain"
)

type CartRepository struct {
	DB *pgxpool.Pool
}

func NewCartRepository(db *pgxpool.Pool) *CartRepository {
	return &CartRepository{DB: db}
}

func (repo *CartRepository) AddToCart(ctx context.Context, userID int64, productVariantID int64, quantity int64) (*int64, error) {

	var cartID int64

	// check whether cart exist
	query := `SELECT id FROM carts WHERE user_id = $1`
	err := repo.DB.QueryRow(ctx, query, userID).Scan(&cartID)
	if err != nil {

		// if cart does not exist, create a new cart
		if err == pgx.ErrNoRows {
			err = repo.DB.QueryRow(ctx, `INSERT INTO carts (user_id) VALUES ($1) RETURNING id`, userID).Scan(&cartID)
			if err != nil {
				return nil, err
			}
		}
	}

	// insert or update cart item
	query = `INSERT INTO cart_items (cart_id, product_variant_id, quantity) VALUES ($1, $2, $3)
	ON CONFLICT (cart_id, product_variant_id)
	 DO UPDATE SET quantity = cart_items.quantity + EXCLUDED.quantity
	 RETURNING cart_id`
	err = repo.DB.QueryRow(ctx, query, cartID, productVariantID, quantity).Scan(&cartID)
	if err != nil {

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23503" {
				switch pgErr.ConstraintName {
				case "cart_items_product_variant_id_fkey":
					return nil, apperror.ErrProductVariantNotFound
				}
			}
		}

		return nil, err
	}
	return &cartID, nil
}

func (repo *CartRepository) GetCartByID(ctx context.Context, cartID int64) (*domain.Cart, error) {

	var cartItems []*domain.CartItem
	var _ domain.CartItem

	query := `
	SELECT ci.product_variant_id,
	p.name as product_name, 
	pv.sku, 
	pv.original_price, 
	pv.sale_price,
	pv.stock, 
	COALESCE((SELECT url FROM product_variant_images WHERE product_variant_id = ci.product_variant_id LIMIT 1), 'null') as image_url, 
	ci.quantity
	FROM cart_items ci
	LEFT JOIN product_variants pv ON ci.product_variant_id = pv.id
	LEFT JOIN products p ON pv.product_id = p.id
	LEFT JOIN product_variant_images pvi ON pv.id = pvi.product_variant_id
	WHERE cart_id = $1
	`
	rows, err := repo.DB.Query(ctx, query, cartID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var totalPrice float64
	for rows.Next() {
		var cartItem domain.CartItem
		err = rows.Scan(&cartItem.ProductVariantID, &cartItem.ProductName, &cartItem.SKU, &cartItem.OriginalPrice,
			&cartItem.SalePrice, &cartItem.Stock, &cartItem.ImageURL, &cartItem.Quantity)
		if err != nil {
			return nil, err
		}

		if cartItem.SalePrice != nil && *cartItem.SalePrice != 0 {
			cartItem.TotalPrice = *cartItem.SalePrice * float64(cartItem.Quantity)
		} else {
			cartItem.TotalPrice = cartItem.OriginalPrice * float64(cartItem.Quantity)
		}
		if cartItem.Stock == 0 {
			cartItem.Status = domain.StatusOutOfStock
			cartItem.Message = "Out of stock"
		} else if cartItem.Stock < int(cartItem.Quantity) && cartItem.Stock > 0 {
			cartItem.Status = domain.StatusLowStock
			cartItem.Message = "Low stock"
		} else {
			cartItem.Status = domain.StatusInStock
			cartItem.Message = "In stock"
		}
		totalPrice += cartItem.TotalPrice
		cartItems = append(cartItems, &cartItem)
	}

	return &domain.Cart{
		Items:          cartItems,
		CartTotalPrice: totalPrice,
	}, nil
}

func (repo *CartRepository) GetCartByUserID(ctx context.Context, userID int64) (*domain.Cart, error) {
	var cartItems []*domain.CartItem
	var _ domain.CartItem
	query := `
	SELECT ci.product_variant_id,
	p.name as product_name, 
	pv.sku, 
	pv.original_price, 
	pv.sale_price,
	pv.stock, 
	COALESCE((SELECT url FROM product_variant_images WHERE product_variant_id = ci.product_variant_id LIMIT 1), 'null') as image_url, 
	ci.quantity
	FROM cart_items ci
	LEFT JOIN product_variants pv ON ci.product_variant_id = pv.id
	LEFT JOIN products p ON pv.product_id = p.id
	LEFT JOIN product_variant_images pvi ON pv.id = pvi.product_variant_id
	INNER JOIN carts c ON ci.cart_id = c.id
	WHERE c.user_id = $1
	`
	rows, err := repo.DB.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var totalPrice float64
	for rows.Next() {
		var cartItem domain.CartItem
		err = rows.Scan(&cartItem.ProductVariantID, &cartItem.ProductName, &cartItem.SKU, &cartItem.OriginalPrice,
			&cartItem.SalePrice, &cartItem.Stock, &cartItem.ImageURL, &cartItem.Quantity)
		if err != nil {
			return nil, err
		}

		if cartItem.SalePrice != nil && *cartItem.SalePrice != 0 {
			cartItem.TotalPrice = *cartItem.SalePrice * float64(cartItem.Quantity)
		} else {
			cartItem.TotalPrice = cartItem.OriginalPrice * float64(cartItem.Quantity)
		}
		if cartItem.Stock == 0 {
			cartItem.Status = domain.StatusOutOfStock
			cartItem.Message = "Out of stock"
		} else if cartItem.Stock < int(cartItem.Quantity) && cartItem.Stock > 0 {
			cartItem.Status = domain.StatusLowStock
			cartItem.Message = "Low stock"
		} else {
			cartItem.Status = domain.StatusInStock
			cartItem.Message = "In stock"
		}
		totalPrice += cartItem.TotalPrice
		cartItems = append(cartItems, &cartItem)
	}

	return &domain.Cart{
		Items:          cartItems,
		CartTotalPrice: totalPrice,
	}, nil
}

func (repo *CartRepository) UpdateCartItemQuantity(ctx context.Context, userID int64, req *domain.UpdateCartItemQuantityRequest) (int64, error) {

	var cartID int64
	query := `UPDATE cart_items ci
	 SET quantity = $1
	 FROM carts c
	 WHERE ci.cart_id = c.id
	 AND c.user_id = $2
	 AND ci.product_variant_id = $3
	RETURNING ci.cart_id`
	err := repo.DB.QueryRow(ctx, query, req.Quantity, userID, req.ProductVariantID).Scan(&cartID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, apperror.ErrCartItemNotFound
		}
		return 0, err
	}
	return cartID, nil
}

func (repo *CartRepository) RemoveCartItem(ctx context.Context, userID int64, productVariantID int64) (int64, error) {
	var cartID int64
	query := `DELETE FROM cart_items ci
	USING carts c
	 WHERE ci.cart_id = c.id
	 AND c.user_id = $1
	 AND ci.product_variant_id = $2
	RETURNING c.id`
	err := repo.DB.QueryRow(ctx, query, userID, productVariantID).Scan(&cartID)
	if err != nil {
		return 0, err
	}
	if err == pgx.ErrNoRows {
		return 0, apperror.ErrCartItemNotFound
	}
	return cartID, nil
}

func (repo *CartRepository) EmptyCart(ctx context.Context, userID int64) error {
	query := `DELETE FROM cart_items ci
	USING carts c
	 WHERE ci.cart_id = c.id
	 AND c.user_id = $1
	RETURNING ci.cart_id`
	cmdTag, err := repo.DB.Exec(ctx, query, userID)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return apperror.ErrCartEmpty
	}
	return nil
}

func (repo *CartRepository) GetCartVariantStocks(ctx context.Context, userID int64) (map[int64]int64, error) {

	cartVariantStocks := make(map[int64]int64)

	query := `SELECT pv.id, pv.stock FROM product_variants pv
	LEFT JOIN cart_items ci ON pv.id = ci.product_variant_id
	LEFT JOIN carts c ON ci.cart_id = c.id
	WHERE c.user_id = $1`

	rows, err := repo.DB.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var variant domain.VariantStock
		if err := rows.Scan(&variant.ProductVariantID, &variant.Stock); err != nil {
			return nil, err
		}
		cartVariantStocks[variant.ProductVariantID] = variant.Stock
	}

	return cartVariantStocks, nil
}
