package repository

import (
	"context"
	"database/sql"
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

// Product operations
// //////////////////////////////////////////////////////////
// //////////////////////////////////////////////////////////
func (repo *ProductRepository) CreateProduct(ctx context.Context, product *domain.CreateProductRequest) (*domain.Product, error) {
	query := `
		INSERT INTO products (name, brand_id, description, category_id, base_price, is_digital, is_active, image_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
		RETURNING id, name, brand_id, description, category_id, base_price, is_digital, is_active, image_url, created_at, updated_at
	`
	var created domain.Product
	err := repo.DB.QueryRow(ctx, query, product.Name, product.BrandID, product.Description, product.CategoryId,
		product.BasePrice, product.IsDigital, product.IsActive, product.ImageURL).Scan(&created.ID, &created.Name, &created.BrandID, &created.Description,
		&created.CategoryId, &created.BasePrice, &created.IsDigital, &created.IsActive, &created.ImageURL, &created.CreatedAt, &created.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &created, nil
}

func (repo *ProductRepository) GetProducts(ctx context.Context, page, limit int64, activeOnly bool) ([]*domain.ProductResponse, int64, error) {

	offset := (page - 1) * limit

	var total int64
	query := `
		SELECT COUNT(*) FROM products
	`
	err := repo.DB.QueryRow(ctx, query).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query = `
		SELECT p.id, p.name, b.name as brand_name, COALESCE(p.description, ''), c.name as category_name, p.base_price, p.is_digital, p.is_active, p.image_url, p.created_at, p.updated_at
		FROM products p
		LEFT JOIN brands b ON p.brand_id = b.id
		LEFT JOIN categories c ON p.category_id = c.id
	`
	if activeOnly {
		query += " WHERE p.is_active = true"
	}
	query += " ORDER BY p.id LIMIT $1 OFFSET $2"
	rows, err := repo.DB.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	products := []*domain.ProductResponse{}
	var p domain.ProductResponse
	for rows.Next() {

		if err := rows.Scan(&p.ID, &p.Name, &p.Brand.ID, &p.Brand.Name, &p.Description, &p.Category.ID, &p.Category.Name,
			&p.BasePrice, &p.IsDigital, &p.IsActive, &p.ImageURL, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, 0, err
		}

		products = append(products, &p)
	}
	if rows.Err() != nil {
		return nil, 0, rows.Err()
	}
	return products, total, nil
}

func (repo *ProductRepository) GetProductByID(ctx context.Context, id int64, activeOnly bool) (*domain.ProductResponse, error) {
	query := `
		SELECT id, name, brand_id,brands.name as brand_name, COALESCE(description, ''), category_id,categories.name as category_name, base_price, is_digital, is_active, image_url, created_at, updated_at
		FROM products LEFT JOIN brands ON products.brand_id = brands.id LEFT JOIN categories ON products.category_id = categories.id
		WHERE id = $1
	`
	if activeOnly {
		query += " AND is_active = true"
	}
	var p domain.ProductResponse

	rows, err := repo.DB.Query(ctx, query, id)

	if err == sql.ErrNoRows {
		fmt.Println("errr : sql.ErrNoRows")
		return nil, fmt.Errorf("PRODUCT_NOT_FOUND")
	}
	if err != nil {
		fmt.Println("errr : ", err.Error())
		return nil, err
	}

	defer rows.Close()

	if rows.Next() {
		err := rows.Scan(&p.ID, &p.Name, &p.Brand.ID, &p.Brand.Name, &p.Description, &p.Category.ID,
			&p.Category.Name, &p.BasePrice, &p.IsDigital, &p.IsActive, &p.ImageURL, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		return &p, nil
	}
	return nil, fmt.Errorf("PRODUCT_NOT_FOUND")
}

func (repo *ProductRepository) UpdateProduct(ctx context.Context, product *domain.UpdateProductRequest) error {
	query := `UPDATE products 
	SET name = COALESCE($1,name), brand_id = COALESCE($2,brand_id), description = COALESCE($3,description),
	 category_id = COALESCE($4,category_id), image_url = COALESCE($5,image_url), base_price = COALESCE($6,base_price),
	 is_digital = COALESCE($7,is_digital), is_active = COALESCE($8,is_active), updated_at = now()
	WHERE id = $9
	`
	cmdTag, err := repo.DB.Exec(ctx, query, product.Name, product.BrandID, product.Description,
		product.CategoryId, product.ImageURL, product.BasePrice, product.IsDigital, product.IsActive, product.ID)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("PRODUCT_NOT_FOUND")
	}
	return nil
}

func (repo *ProductRepository) DeleteProduct(ctx context.Context, id int64) error {
	cmdTag, err := repo.DB.Exec(ctx, `DELETE FROM products WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("PRODUCT_NOT_FOUND")
	}
	return nil
}
