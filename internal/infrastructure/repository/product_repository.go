package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tharunn0/E-Commerce-Go/internal/domain"
)

type ProductRepository struct {
	DB *pgxpool.Pool
}

func NewProductRepository(db *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{DB: db}
}

// Brand operations
// //////////////////////////////////////////////////////////
func (repo *ProductRepository) CreateBrand(ctx context.Context, createBrandRequest *domain.CreateBrandRequest) (*domain.Brand, error) {
	query := `
		INSERT INTO brands (name, description, logo_url)
		VALUES ($1, $2, $3)
		RETURNING id, name, description, logo_url, is_active, created_at, updated_at
	`
	var created domain.Brand
	err := repo.DB.QueryRow(ctx, query, createBrandRequest.Name, createBrandRequest.Description, createBrandRequest.LogoURL).Scan(
		&created.ID, &created.Name, &created.Description, &created.LogoURL, &created.IsActive, &created.CreatedAt, &created.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &created, nil
}

func (repo *ProductRepository) GetAllBrands(ctx context.Context, activeOnly bool) ([]*domain.Brand, error) {
	query := `
		SELECT id, name, description, logo_url, is_active, created_at, updated_at
		FROM brands
	`
	if activeOnly {
		query += " WHERE is_active = true"
	}
	rows, err := repo.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	brands := []*domain.Brand{}
	for rows.Next() {
		var b domain.Brand
		err := rows.Scan(&b.ID, &b.Name, &b.Description, &b.LogoURL, &b.IsActive, &b.CreatedAt, &b.UpdatedAt)
		if err != nil {
			return nil, err
		}
		brands = append(brands, &b)
	}

	return brands, nil
}

func (repo *ProductRepository) GetBrandByID(ctx context.Context, id int64, activeOnly bool) (*domain.Brand, error) {
	query := `
		SELECT id, name, description, logo_url, is_active, created_at, updated_at
		FROM brands
		WHERE id = $1
	`
	if activeOnly {
		query += " AND is_active = true"
	}
	var brand domain.Brand
	err := repo.DB.QueryRow(ctx, query, id).Scan(&brand.ID, &brand.Name, &brand.Description, &brand.LogoURL, &brand.IsActive, &brand.CreatedAt, &brand.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &brand, nil
}

func (repo *ProductRepository) UpdateBrand(ctx context.Context, brand *domain.Brand) (*domain.Brand, error) {
	query := `
		UPDATE brands SET name = $1, description = $2, logo_url = $3, is_active = $4, updated_at = now() WHERE id = $5
		RETURNING id, name, description, logo_url, is_active, created_at, updated_at
	`
	var updated domain.Brand
	err := repo.DB.QueryRow(ctx, query, brand.Name, brand.Description, brand.LogoURL, brand.IsActive, brand.ID).Scan(&updated.ID, &updated.Name, &updated.Description, &updated.LogoURL, &updated.IsActive, &updated.CreatedAt, &updated.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

func (repo *ProductRepository) DeleteBrand(ctx context.Context, id int64) error {
	cmdTag, err := repo.DB.Exec(ctx, `DELETE FROM brands WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("BRAND_NOT_FOUND")
	}
	return nil
}
