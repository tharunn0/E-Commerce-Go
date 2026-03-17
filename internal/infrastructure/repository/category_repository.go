package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/catalog"
)

type CategoryRepository struct {
	DB *pgxpool.Pool
}

func NewCategoryRepository(db *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{DB: db}
}

func (repo *CategoryRepository) CreateCategory(ctx context.Context, category *catalog.Category) (*catalog.Category, error) {
	query := `
		INSERT INTO categories (name, description, parent_category_id, image_url)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, description, parent_category_id, image_url, is_active, created_at, updated_at
	`
	var created catalog.Category
	err := repo.DB.QueryRow(ctx, query, category.Name, category.Description, category.ParentCategoryID, category.ImageURL).Scan(
		&created.ID, &created.Name, &created.Description, &created.ParentCategoryID, &created.ImageURL, &created.IsActive, &created.CreatedAt, &created.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &created, nil
}

func (repo *CategoryRepository) UpdateCategory(ctx context.Context, category *catalog.Category) (*catalog.Category, error) {
	query := `
		UPDATE categories SET name = $1, description = $2, parent_category_id = $3, image_url = $4, updated_at = now()
		WHERE id = $5
		RETURNING id, name, description, parent_category_id, image_url, is_active, created_at, updated_at
	`
	var updated catalog.Category
	err := repo.DB.QueryRow(ctx, query, category.Name, category.Description, category.ParentCategoryID, category.ImageURL, category.ID).Scan(
		&updated.ID, &updated.Name, &updated.Description, &updated.ParentCategoryID, &updated.ImageURL, &updated.IsActive, &updated.CreatedAt, &updated.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

func (repo *CategoryRepository) ChangeCategoryStatus(ctx context.Context, id int64, isActive bool) (*catalog.Category, error) {
	query := `UPDATE categories SET is_active = $1 WHERE id = $2 RETURNING id, name, description, parent_category_id, image_url, is_active, created_at, updated_at`
	var updated catalog.Category
	err := repo.DB.QueryRow(ctx, query, isActive, id).Scan(
		&updated.ID,
		&updated.Name,
		&updated.Description,
		&updated.ParentCategoryID,
		&updated.ImageURL,
		&updated.IsActive,
		&updated.CreatedAt, &updated.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &updated, nil
}

func (repo *CategoryRepository) DeleteCategory(ctx context.Context, id int64) error {
	cmdTag, err := repo.DB.Exec(ctx, `DELETE FROM categories WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("category not found")
	}
	return nil
}

func (repo *CategoryRepository) GetCategoryByID(ctx context.Context, id int64, activeOnly bool) (*catalog.Category, error) {
	var category catalog.Category
	query := `SELECT id, name, description, parent_category_id, image_url, created_at, updated_at FROM categories WHERE id = $1`
	if activeOnly {
		query += " AND is_active = true"
	}
	err := repo.DB.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&category.ID,
		&category.Name,
		&category.Description,
		&category.ParentCategoryID,
		&category.ImageURL,
		&category.CreatedAt,
		&category.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (repo *CategoryRepository) GetAllCategories(ctx context.Context, activeOnly bool) ([]*catalog.Category, error) {
	query := `SELECT id, name, description, parent_category_id, image_url, is_active, created_at, updated_at FROM categories`
	if activeOnly {
		query += " WHERE is_active = true"
	}
	rows, err := repo.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := []*catalog.Category{}
	for rows.Next() {
		var category catalog.Category
		err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.Description,
			&category.ParentCategoryID,
			&category.ImageURL,
			&category.IsActive,
			&category.CreatedAt,
			&category.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		categories = append(categories, &category)
	}
	return categories, nil
}

