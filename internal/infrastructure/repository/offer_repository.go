package repository

import (
	"context"
	"log"

	"github.com/tharunn0/E-Commerce-Go/internal/apperror"
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

	log.Println("Offer id : ", offerID)

	switch req.Scope {
	case "product":
		for _, productID := range req.ScopeIDs {
			_, err = tx.Exec(ctx, "INSERT INTO product_offers (offer_id, product_id) VALUES ($1, $2)", offerID, productID)
			if err != nil {
				log.Println("Failed to create product offer", "[id :", productID, "] [error : ", err, "]")
				return apperror.ErrProductDoesNotExist
			}
		}
	case "category":
		for _, categoryID := range req.ScopeIDs {
			_, err = tx.Exec(ctx, "INSERT INTO category_offers (offer_id, category_id) VALUES ($1, $2)", offerID, categoryID)
			if err != nil {
				log.Println("Failed to create category offer", "[id :", categoryID, "] [error : ", err, "]")
				return apperror.ErrCategoryDoesNotExist
			}
		}
	default:
		log.Println("Invalid scope for offer", "[scope : ", req.Scope, "]")
		return nil
	}

	return tx.Commit(ctx)
}

func (repo *OfferRepository) GetAllCategoryOffers(ctx context.Context) ([]*domain.Offer, error) {

	log.Println("GetAllCategoryOffers called")

	query := `
		SELECT
			o.id,
			o.name,
			o.description,
			o.discount_type,
			o.discount_value,
			o.start_date,
			o.end_date,
			o.is_active,
			o.created_at,
			COALESCE(
				ARRAY_AGG(co.category_id)
					FILTER (WHERE co.category_id IS NOT NULL),
				'{}'
			) AS eligible_ids
		FROM offers o
		INNER JOIN category_offers co
			ON co.offer_id = o.id
		GROUP BY o.id
		ORDER BY o.created_at DESC
	`

	rows, err := repo.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	offers := make([]*domain.Offer, 0)

	for rows.Next() {
		offer := new(domain.Offer)

		err := rows.Scan(
			&offer.ID,
			&offer.Name,
			&offer.Description,
			&offer.DiscountType,
			&offer.DiscountValue,
			&offer.StartDate,
			&offer.EndDate,
			&offer.IsActive,
			&offer.CreatedAt,
			&offer.EligibleIds,
		)
		if err != nil {
			return nil, err
		}

		offer.Scope = "category"
		offers = append(offers, offer)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return offers, nil
}

func (repo *OfferRepository) GetCategoryOffers(ctx context.Context, ids []int64) ([]*domain.Offer, error) {

	//     ID            int64     `json:"id"`
	// Name          string    `json:"name"`
	// Description   string    `json:"description"`
	// DiscountType  string    `json:"discount_type"`
	// DiscountValue float64   `json:"discount_value"`
	// StartDate     time.Time `json:"start_date"`
	// EndDate       time.Time `json:"end_date"`
	// IsActive      bool      `json:"is_active"`
	// CreatedAt     time.Time `json:"created_at"`

	query := `SELECT o.id, o.name, o.description, o.discount_type, o.discount_value, o.start_date, 
	o.end_date, o.is_active, o.created_at, co.category_id
	FROM offers o
	JOIN category_offers co ON co.offer_id = o.id
	WHERE
    co.category_id = ANY($1)
    AND o.is_active = true
    AND o.start_date <= NOW()
    AND o.end_date >= NOW();`

	var offers []*domain.Offer
	rows, err := repo.db.Query(ctx, query, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var offer domain.Offer
		err := rows.Scan(
			&offer.ID,
			&offer.Name,
			&offer.Description,
			&offer.DiscountType,
			&offer.DiscountValue,
			&offer.StartDate,
			&offer.EndDate,
			&offer.IsActive,
			&offer.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		offer.Scope = "category"
		offers = append(offers, &offer)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return offers, nil
}

func (repo *OfferRepository) GetAllProductOffers(ctx context.Context) ([]*domain.Offer, error) {

	log.Println("GetAllProductOffers called")

	query := `
		SELECT
			o.id,
			o.name,
			o.description,
			o.discount_type,
			o.discount_value,
			o.start_date,
			o.end_date,
			o.is_active,
			o.created_at,
			COALESCE(
				ARRAY_AGG(po.product_id)
					FILTER (WHERE po.product_id IS NOT NULL),
				'{}'
			) AS eligible_ids
		FROM offers o
		INNER JOIN product_offers po
			ON po.offer_id = o.id
		GROUP BY o.id
		ORDER BY o.created_at DESC
	`

	rows, err := repo.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	offers := make([]*domain.Offer, 0)

	for rows.Next() {
		offer := new(domain.Offer)

		err := rows.Scan(
			&offer.ID,
			&offer.Name,
			&offer.Description,
			&offer.DiscountType,
			&offer.DiscountValue,
			&offer.StartDate,
			&offer.EndDate,
			&offer.IsActive,
			&offer.CreatedAt,
			&offer.EligibleIds,
		)

		offer.Scope = "product"
		if err != nil {
			return nil, err
		}

		log.Println("offer : ", offer)

		offers = append(offers, offer)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return offers, nil
}

func (repo *OfferRepository) GetProductOffers(ctx context.Context, ids []int64) ([]*domain.Offer, error) {

	query := `SELECT o.id, o.name, o.description, o.discount_type, o.discount_value, o.start_date, 
	o.end_date, o.is_active, o.created_at, po.product_id
	FROM offers o
	JOIN product_offers po ON po.offer_id = o.id
	WHERE
    po.product_id = ANY($1)
    AND o.is_active = true
    AND o.start_date <= NOW()
    AND o.end_date >= NOW();`

	var offers []*domain.Offer
	rows, err := repo.db.Query(ctx, query, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var offer domain.Offer
		err := rows.Scan(
			&offer.ID,
			&offer.Name,
			&offer.Description,
			&offer.DiscountType,
			&offer.DiscountValue,
			&offer.StartDate,
			&offer.EndDate,
			&offer.IsActive,
			&offer.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		offer.Scope = "product"
		offers = append(offers, &offer)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return offers, nil
}

// func (repo *OfferRepository) GetOfferByID(ctx context.Context, id int64) (*domain.Offer, error) {

// 	var query string

// 	var offer domain.Offer
// 	err := repo.db.QueryRow(ctx, query, id).Scan(
// 		&offer.ID,
// 		&offer.Name,
// 		&offer.Description,
// 		&offer.DiscountType,
// 		&offer.DiscountValue,
// 		&offer.StartDate,
// 		&offer.EndDate,
// 		&offer.IsActive,
// 		&offer.CreatedAt,
// 	)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &offer, nil
// }
