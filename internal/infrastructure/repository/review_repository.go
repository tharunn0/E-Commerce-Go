package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tharunn0/E-Commerce-Go/internal/apperror"
	"github.com/tharunn0/E-Commerce-Go/internal/domain"
)

type ReviewRepository struct {
	DB *pgxpool.Pool
}

func NewReviewRepository(db *pgxpool.Pool) *ReviewRepository {
	return &ReviewRepository{DB: db}
}

func (r *ReviewRepository) CreateReview(ctx context.Context, review *domain.CreateReviewRequest) error {

	// check if user has ordered the product
	query := `SELECT EXISTS (
		SELECT 1 from orders o
		JOIN order_items oi ON o.id = oi.order_id
		JOIN product_variants pv ON pv.id = oi.product_variant_id
		WHERE o.user_id = $1 AND pv.product_id = $2 AND o.status = 'delivered'
	)`

	var exists bool
	err := r.DB.QueryRow(ctx, query, review.UserID, review.ProductID).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return apperror.ErrProductNotOrdered
	}

	// check if review already exists
	query = `SELECT EXISTS (
		SELECT 1 FROM reviews WHERE user_id = $1 AND product_id = $2
	)`
	err = r.DB.QueryRow(ctx, query, review.UserID, review.ProductID).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return apperror.ErrReviewAlreadyExists
	}

	// insert review
	query = `INSERT INTO reviews (user_id, product_id, title, body, rating)
				VALUES ($1, $2, $3, $4, $5)`

	cmdTag, err := r.DB.Exec(ctx, query, review.UserID, review.ProductID, review.Title, review.Description, review.Rating)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return apperror.ErrFailedToCreateReview
	}

	return nil
}

func (r *ReviewRepository) DeleteReview(ctx context.Context, id int64) error {
	return nil
}

func (r *ReviewRepository) GetProductReviews(ctx context.Context, productID int64, filter *domain.ReviewFilter) (*domain.ProductReviews, error) {

	query := `SELECT 
		r.id,
		u.first_name,
		u.last_name,
		r.title,
		r.body,
		r.rating,
		r.created_at,
		r.updated_at
		FROM reviews r
		JOIN users u ON r.user_id = u.id
		WHERE r.product_id = $1`

	var firstname, lastname string
	rows, err := r.DB.Query(ctx, query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []domain.Review
	for rows.Next() {
		var review domain.Review
		if err := rows.Scan(&review.ID, &firstname, &lastname, &review.Title, &review.Description, &review.Rating, &review.CreatedAt, &review.UpdatedAt); err != nil {
			return nil, err
		}
		review.User = firstname + " " + lastname
		reviews = append(reviews, review)
	}

	reviewResponse := domain.ProductReviews{
		ProductID:     productID,
		ProductName:   "",
		AverageRating: 0,
		TotalReviews:  int64(len(reviews)),
		Reviews:       reviews,
	}

	query = `SELECT p.name FROM products p JOIN reviews r ON p.id = r.product_id WHERE r.product_id = $1`
	err = r.DB.QueryRow(ctx, query, productID).Scan(&reviewResponse.ProductName)
	if err != nil {
		return nil, err
	}

	for _, review := range reviews {
		reviewResponse.AverageRating += review.Rating
	}
	reviewResponse.AverageRating /= float64(len(reviews))

	return &reviewResponse, nil
}
