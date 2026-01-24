package repository

import (
	"context"

	"github.com/tharunn0/E-Commerce-Go/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OfferRepository struct {
	db *pgxpool.Pool
}

func NewOfferRepository(db *pgxpool.Pool) *OfferRepository {
	return &OfferRepository{db: db}
}

func (repo *OfferRepository) CreateOffer(ctx context.Context, req *domain.CreateOfferRequest) error {
	tx, err := repo.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var offerID int64
	query := `INSERT INTO offers 
		(name, description, discount_type, discount_value, start_date, end_date, is_active) 
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`

	err = tx.QueryRow(ctx, query,
		req.Name,
		req.Description,
		req.DiscountType,
		req.DiscountValue,
		req.StartDate,
		req.EndDate,
		true,
	).Scan(&offerID)
	if err != nil {
		return err
	}

	if req.Scope == "product" {
		for _, productID := range req.ScopeIDs {
			_, err = tx.Exec(ctx, "INSERT INTO product_offers (offer_id, product_id) VALUES ($1, $2)", offerID, productID)
			if err != nil {
				return err
			}
		}
	} else if req.Scope == "category" {
		for _, categoryID := range req.ScopeIDs {
			_, err = tx.Exec(ctx, "INSERT INTO category_offers (offer_id, category_id) VALUES ($1, $2)", offerID, categoryID)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit(ctx)
}
