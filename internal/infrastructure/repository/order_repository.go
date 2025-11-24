package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tharunn0/E-Commerce-Go/internal/domain"
)

type OrderRepository struct {
	DB *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) OrderRepository {
	return OrderRepository{DB: db}
}

func (repo OrderRepository) ValidateCart(ctx context.Context, userID int64) (map[int64]int64, error) {

	fmt.Println("validate cart called in repository")

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
