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

// func (repo *OfferRepository) GetAllCategoryOffers(ctx context.Context) ([]*domain.Offer, error) {

// 	log.Println("GetAllCategoryOffers called")

// 	query := `
// 		SELECT
// 			o.id,
// 			o.name,
// 			o.description,
// 			o.discount_type,
// 			o.discount_value,
// 			o.start_date,
// 			o.end_date,
// 			o.is_active,
// 			o.created_at,
// 			COALESCE(
// 				ARRAY_AGG(co.category_id)
// 					FILTER (WHERE co.category_id IS NOT NULL),
// 				'{}'
// 			) AS eligible_ids
// 		FROM offers o
// 		INNER JOIN category_offers co
// 			ON co.offer_id = o.id
// 		GROUP BY o.id
// 		ORDER BY o.created_at DESC
// 	`

// 	rows, err := repo.db.Query(ctx, query)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()

// 	offers := make([]*domain.Offer, 0)

// 	for rows.Next() {
// 		offer := new(domain.Offer)

// 		err := rows.Scan(
// 			&offer.ID,
// 			&offer.Name,
// 			&offer.Description,
// 			&offer.DiscountType,
// 			&offer.DiscountValue,
// 			&offer.StartDate,
// 			&offer.EndDate,
// 			&offer.IsActive,
// 			&offer.CreatedAt,
// 			&offer.EligibleIds,
// 		)
// 		if err != nil {
// 			return nil, err
// 		}

// 		offer.Scope = "category"
// 		offers = append(offers, offer)
// 	}

// 	if err := rows.Err(); err != nil {
// 		return nil, err
// 	}

// 	return offers, nil
// }

// func (repo *OfferRepository) GetCategoryOffers(ctx context.Context, ids []int64) ([]*domain.Offer, error) {

// 	//     ID            int64     `json:"id"`
// 	// Name          string    `json:"name"`
// 	// Description   string    `json:"description"`
// 	// DiscountType  string    `json:"discount_type"`
// 	// DiscountValue float64   `json:"discount_value"`
// 	// StartDate     time.Time `json:"start_date"`
// 	// EndDate       time.Time `json:"end_date"`
// 	// IsActive      bool      `json:"is_active"`
// 	// CreatedAt     time.Time `json:"created_at"`

// 	query := `SELECT o.id, o.name, o.description, o.discount_type, o.discount_value, o.start_date,
// 	o.end_date, o.is_active, o.created_at, co.category_id
// 	FROM offers o
// 	JOIN category_offers co ON co.offer_id = o.id
// 	WHERE
//     co.category_id = ANY($1)
//     AND o.is_active = true
//     AND o.start_date <= NOW()
//     AND o.end_date >= NOW();`

// 	var offers []*domain.Offer
// 	rows, err := repo.db.Query(ctx, query, ids)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()
// 	for rows.Next() {
// 		var offer domain.Offer
// 		err := rows.Scan(
// 			&offer.ID,
// 			&offer.Name,
// 			&offer.Description,
// 			&offer.DiscountType,
// 			&offer.DiscountValue,
// 			&offer.StartDate,
// 			&offer.EndDate,
// 			&offer.IsActive,
// 			&offer.CreatedAt,
// 		)
// 		if err != nil {
// 			return nil, err
// 		}
// 		offer.Scope = "category"
// 		offers = append(offers, &offer)
// 	}
// 	if err := rows.Err(); err != nil {
// 		return nil, err
// 	}

// 	return offers, nil
// }

// func (repo *OfferRepository) GetAllProductOffers(ctx context.Context) ([]*domain.Offer, error) {

// 	log.Println("GetAllProductOffers called")

// 	query := `
// 		SELECT
// 			o.id,
// 			o.name,
// 			o.description,
// 			o.discount_type,
// 			o.discount_value,
// 			o.start_date,
// 			o.end_date,
// 			o.is_active,
// 			o.created_at,
// 			COALESCE(
// 				ARRAY_AGG(po.product_id)
// 					FILTER (WHERE po.product_id IS NOT NULL),
// 				'{}'
// 			) AS eligible_ids
// 		FROM offers o
// 		INNER JOIN product_offers po
// 			ON po.offer_id = o.id
// 		GROUP BY o.id
// 		ORDER BY o.created_at DESC
// 	`

// 	rows, err := repo.db.Query(ctx, query)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()

// 	offers := make([]*domain.Offer, 0)

// 	for rows.Next() {
// 		offer := new(domain.Offer)

// 		err := rows.Scan(
// 			&offer.ID,
// 			&offer.Name,
// 			&offer.Description,
// 			&offer.DiscountType,
// 			&offer.DiscountValue,
// 			&offer.StartDate,
// 			&offer.EndDate,
// 			&offer.IsActive,
// 			&offer.CreatedAt,
// 			&offer.EligibleIds,
// 		)

// 		offer.Scope = "product"
// 		if err != nil {
// 			return nil, err
// 		}

// 		log.Println("offer : ", offer)

// 		offers = append(offers, offer)
// 	}

// 	if err := rows.Err(); err != nil {
// 		return nil, err
// 	}

// 	return offers, nil
// }

// func (repo *OfferRepository) GetProductOffers(ctx context.Context, ids []int64) ([]*domain.Offer, error) {

// 	query := `SELECT o.id, o.name, o.description, o.discount_type, o.discount_value, o.start_date,
// 	o.end_date, o.is_active, o.created_at, po.product_id
// 	FROM offers o
// 	JOIN product_offers po ON po.offer_id = o.id
// 	WHERE
//     po.product_id = ANY($1)
//     AND o.is_active = true
//     AND o.start_date <= NOW()
//     AND o.end_date >= NOW();`

// 	var offers []*domain.Offer
// 	rows, err := repo.db.Query(ctx, query, ids)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()
// 	for rows.Next() {
// 		var offer domain.Offer
// 		err := rows.Scan(
// 			&offer.ID,
// 			&offer.Name,
// 			&offer.Description,
// 			&offer.DiscountType,
// 			&offer.DiscountValue,
// 			&offer.StartDate,
// 			&offer.EndDate,
// 			&offer.IsActive,
// 			&offer.CreatedAt,
// 		)
// 		if err != nil {
// 			return nil, err
// 		}
// 		offer.Scope = "product"
// 		offers = append(offers, &offer)
// 	}
// 	if err := rows.Err(); err != nil {
// 		return nil, err
// 	}

// 	return offers, nil
// }

func (repo *OfferRepository) GetAllActiveOffers(
	ctx context.Context,
	productIDs []int64,
	categoryIDs []int64,
) ([]*domain.Offer, error) {

	// 1️⃣ Fetch relevant offers
	query := `
	SELECT DISTINCT o.id, o.name, o.description,
	       o.discount_type, o.discount_value,
	       o.start_date, o.end_date,
	       o.is_active, o.created_at
	FROM offers o
	LEFT JOIN product_offers po ON po.offer_id = o.id
	LEFT JOIN category_offers co ON co.offer_id = o.id
	WHERE o.is_active = true
	  AND now() BETWEEN o.start_date AND o.end_date
	  AND (
	       po.product_id = ANY($1)
	    OR co.category_id = ANY($2)
	  );
	`

	rows, err := repo.db.Query(ctx, query, productIDs, categoryIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	offerMap := make(map[int64]*domain.Offer)
	var offerIDs []int64

	for rows.Next() {
		var o domain.Offer
		if err := rows.Scan(
			&o.ID,
			&o.Name,
			&o.Description,
			&o.DiscountType,
			&o.DiscountValue,
			&o.StartDate,
			&o.EndDate,
			&o.IsActive,
			&o.CreatedAt,
		); err != nil {
			return nil, err
		}

		offerMap[o.ID] = &o
		offerIDs = append(offerIDs, o.ID)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(offerIDs) == 0 {
		return []*domain.Offer{}, nil
	}

	// 2️⃣ Load product targets
	productRows, err := repo.db.Query(ctx, `
		SELECT offer_id, product_id
		FROM product_offers
		WHERE offer_id = ANY($1);
	`, offerIDs)
	if err != nil {
		return nil, err
	}
	defer productRows.Close()

	for productRows.Next() {
		var offerID, productID int64
		if err := productRows.Scan(&offerID, &productID); err != nil {
			return nil, err
		}
		offerMap[offerID].ProductIDs =
			append(offerMap[offerID].ProductIDs, productID)
	}

	// 3️⃣ Load category targets
	categoryRows, err := repo.db.Query(ctx, `
		SELECT offer_id, category_id
		FROM category_offers
		WHERE offer_id = ANY($1);
	`, offerIDs)
	if err != nil {
		return nil, err
	}
	defer categoryRows.Close()

	for categoryRows.Next() {
		var offerID, categoryID int64
		if err := categoryRows.Scan(&offerID, &categoryID); err != nil {
			return nil, err
		}
		offerMap[offerID].CategoryIDs =
			append(offerMap[offerID].CategoryIDs, categoryID)
	}

	// 4️⃣ Flatten map → slice
	offers := make([]*domain.Offer, 0, len(offerMap))
	for _, o := range offerMap {
		offers = append(offers, o)
	}

	return offers, nil
}
