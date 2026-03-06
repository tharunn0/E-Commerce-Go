package repository

import (
	"context"
	"errors"
	"log"

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

func (repo *CartRepository) AddToCart(ctx context.Context, userID int64, productVariantID int64, addQty int64) (*int64, error) {

	var cartID int64

	query := `SELECT id FROM carts WHERE user_id = $1`
	err := repo.DB.QueryRow(ctx, query, userID).Scan(&cartID)
	if err != nil {
		return nil, err
	}

	tx, err := repo.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}

	// ensure rollback on error
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// check if enough stock is available
	query = `SELECT stock FROM product_variants WHERE id = $1`
	var stock int64
	err = repo.DB.QueryRow(ctx, query, productVariantID).Scan(&stock)
	if err != nil {
		return nil, err
	}
	if stock < addQty {
		return nil, apperror.ErrNotEnoughStock
	}

	// insert or update cart item
	query = `
		INSERT INTO cart_items (cart_id, product_variant_id, quantity)
		VALUES ($1, $2, $3)
		ON CONFLICT (cart_id, product_variant_id)
		DO UPDATE
		SET quantity = $3
		RETURNING quantity
	`

	var finalQty int64
	err = tx.QueryRow(ctx, query, cartID, productVariantID, addQty).Scan(&finalQty)
	if err != nil {

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23503" {
				if pgErr.ConstraintName == "cart_items_product_variant_id_fkey" {
					return nil, apperror.ErrProductVariantNotFound
				}
			}
		}

		return nil, err
	}

	// commit transaction
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &cartID, nil
}

func (repo *CartRepository) GetCartByID(ctx context.Context, cartID int64) (*domain.Cart, error) {

	log.Println("GetCartByID called")

	var cartItems []*domain.CartItem
	var _ domain.CartItem

	query := `
	SELECT ci.product_variant_id,
	p.id,
	c.id,
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
	LEFT JOIN categories c ON p.category_id = c.id
	WHERE ci.cart_id = $1
	`
	rows, err := repo.DB.Query(ctx, query, cartID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var totalPrice float64
	for rows.Next() {
		var cartItem domain.CartItem
		err = rows.Scan(&cartItem.ProductVariantID, &cartItem.ProductID, &cartItem.CategoryID, &cartItem.ProductName, &cartItem.SKU, &cartItem.OriginalPrice,
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

	log.Println("cart items, ", cartItems)

	return &domain.Cart{
		Items:          cartItems,
		CartTotalPrice: totalPrice,
	}, nil
}

func (repo *CartRepository) GetCartByUserID(ctx context.Context, userID int64) (*domain.Cart, error) {

	log.Println("GetCartByUserID called	")

	var cartItems []*domain.CartItem
	var _ domain.CartItem
	query := `
	SELECT
	ci.product_variant_id,
	p.id,
	ct.id,
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
	LEFT JOIN categories ct ON p.category_id = ct.id
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
		err = rows.Scan(&cartItem.ProductVariantID, &cartItem.ProductID, &cartItem.CategoryID, &cartItem.ProductName, &cartItem.SKU, &cartItem.OriginalPrice,
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

	log.Println("cart items, ", cartItems)

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

func (repo *CartRepository) GetCartVariantInfo(ctx context.Context, userID int64) (map[int64]domain.VariantInfo, error) {

	cartVariantStocks := make(map[int64]domain.VariantInfo)

	query := `SELECT pv.id, p.name, pv.sku , pv.stock FROM product_variants pv
	LEFT JOIN cart_items ci ON pv.id = ci.product_variant_id
	LEFT JOIN carts c ON ci.cart_id = c.id
	LEFT JOIN products p ON pv.product_id = p.id
	WHERE c.user_id = $1`

	rows, err := repo.DB.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var variant domain.VariantInfo
		var pvId int64
		if err := rows.Scan(&pvId, &variant.ProductName, &variant.SKU, &variant.Stock); err != nil {
			return nil, err
		}
		cartVariantStocks[pvId] = variant
	}

	return cartVariantStocks, nil
}
