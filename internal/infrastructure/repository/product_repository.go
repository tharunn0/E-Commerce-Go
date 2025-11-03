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

// Variant operations
// ////////////////////////////////////////////////////////
// ////////////////////////////////////////////////////////
func (repo *ProductRepository) CreateProductVariant(ctx context.Context, req *domain.CreateProductVariantRequest) (*domain.ProductVariantResponse, error) {

	var created domain.ProductVariantResponse
	var _ domain.AttributeValueResponse
	// start transaction
	tx, err := repo.DB.Begin(ctx)
	if err != nil {
		fmt.Println("transaction failed: ", err.Error())
		return nil, err
	}
	defer tx.Rollback(ctx)

	// insert product variant
	query := `
		INSERT INTO product_variants (product_id, sku, price_difference, stock)
		VALUES ($1, $2, $3, $4)
		RETURNING id, product_id, sku, price_difference, stock, created_at`
	err = tx.QueryRow(ctx, query, req.ProductID, req.SKU, req.PriceDifference, req.Stock).Scan(
		&created.ID, &created.BaseProduct.ID, &created.SKU, &created.PriceDifference, &created.Stock, &created.CreatedAt)
	if err != nil {
		fmt.Println("insertion into product_variants table failed: ", err.Error())
		return nil, err
	}

	// insert variant images
	for i := range req.Images {
		query = `
			INSERT INTO product_variant_images (product_variant_id, url)
			VALUES ($1, $2)
		`
		cmdTag, err := tx.Exec(ctx, query, created.ID, req.Images[i])
		if err != nil {
			fmt.Println("insertion into product_variant_images table failed: ", err.Error())
			return nil, err
		}
		if cmdTag.RowsAffected() == 0 {
			return nil, fmt.Errorf("PRODUCT_VARIANT_IMAGE_INSERTION_FAILED")
		}
	}

	// insert variant attributes
	query = `
		INSERT INTO product_variant_attributes (product_variant_id, attribute_id, attribute_value_id)
		VALUES ($1, $2, $3) RETURNING id
	`
	for i := range req.VariantAttributes {
		var attrVal domain.AttributeValueResponse
		// type AttributeValueResponse struct {
		// 	ID        int64  `json:"id"`
		// 	Attribute string `json:"attribute"`
		// 	Value     string `json:"value"`
		// }
		err = tx.QueryRow(ctx,
			query, created.ID,
			req.VariantAttributes[i].AttributeID,
			req.VariantAttributes[i].AttributeValueID).Scan(
			&attrVal.ID)
		if err != nil {
			fmt.Println("insertion into product_variant_attributes table failed: ", err.Error())
			return nil, err
		}
		fmt.Println("attrVal : ", attrVal)
		created.Attributes = append(created.Attributes, attrVal)
	}

	err = tx.Commit(ctx)
	if err != nil {
		fmt.Println("commit failed: ", err.Error())
		return nil, err
	}

	fmt.Println("transaction committed successfully")

	// get base product
	query = `SELECT p.id,p.name,b.id,b.name,p.base_price FROM products p
	LEFT JOIN brands b ON p.brand_id = b.id
	WHERE p.id = $1
	`
	err = repo.DB.QueryRow(ctx, query, req.ProductID).Scan(
		&created.BaseProduct.ID,
		&created.BaseProduct.Name,
		&created.BaseProduct.Brand.ID,
		&created.BaseProduct.Brand.Name,
		&created.BaseProduct.BasePrice)
	if err != nil {
		fmt.Println("insertion into products table failed: ", err.Error())
		return nil, err
	}

	// get total price
	created.TotalPrice = created.BaseProduct.BasePrice + created.PriceDifference

	// get attribute values
	query = `
		SELECT a.name,av.value FROM product_variant_attributes pva 
		LEFT JOIN attributes a ON pva.attribute_id = a.id 
		LEFT JOIN attribute_values av ON pva.attribute_value_id = av.id
		WHERE pva.id = $1
	`

	for i := range req.VariantAttributes {
		err = repo.DB.QueryRow(ctx, query, created.Attributes[i].ID).Scan(
			&created.Attributes[i].Attribute,
			&created.Attributes[i].Value)
		if err != nil {
			fmt.Println("getting attribute values failed: ", err.Error())
			return nil, err
		}
	}
	return &created, nil
}

func (repo *ProductRepository) GetProductVariantByID(ctx context.Context, id int64, activeOnly bool) (*domain.ProductVariantResponse, error) {
	var created domain.ProductVariantResponse
	query := `
		SELECT id, product_id, sku, price_difference, stock, created_at FROM product_variants WHERE id = $1
	`
	if activeOnly {
		query += " AND is_active = true"
	}
	err := repo.DB.QueryRow(ctx, query, id).Scan(&created.ID, &created.BaseProduct.ID, &created.SKU, &created.PriceDifference, &created.Stock, &created.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}

	// get variant images
	query = `
		SELECT url FROM product_variant_images WHERE product_variant_id = $1
	`
	rows, err := repo.DB.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var image string
		err := rows.Scan(&image)
		if err != nil {
			return nil, err
		}
		created.VariantImages = append(created.VariantImages, image)
	}

	// get variant attributes
	query = `
		SELECT a.name,av.value FROM product_variant_attributes pva 
		LEFT JOIN attributes a ON pva.attribute_id = a.id 
		LEFT JOIN attribute_values av ON pva.attribute_value_id = av.id
		WHERE pva.id = $1
	`
	for i := range created.Attributes {
		var attrVal domain.AttributeValueResponse
		err = repo.DB.QueryRow(ctx, query, created.Attributes[i].ID).Scan(
			&attrVal.ID,
			&attrVal.Attribute,
			&attrVal.Value)
		if err != nil {
			return nil, err
		}
		created.Attributes = append(created.Attributes, attrVal)
	}

	return &created, nil
}

// Attribute operations
// //////////////////////////////////////////////////////////
// //////////////////////////////////////////////////////////
func (repo *ProductRepository) CreateAttribute(ctx context.Context, attribute *domain.CreateAttributeRequest) (*domain.Attribute, error) {

	var created domain.Attribute
	tx, err := repo.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	query := `
		INSERT INTO attributes (name, data_type)
		VALUES ($1, $2)
		RETURNING id, name, data_type, is_active, created_at
	`
	err = tx.QueryRow(ctx, query, attribute.Name, attribute.DataType).Scan(
		&created.ID, &created.Name, &created.DataType, &created.IsActive, &created.CreatedAt)
	if err != nil {
		fmt.Println("insertion into attributes table failed: ", err.Error())
		return nil, err
	}

	for _, value := range attribute.Values {
		var val domain.AttributeValue

		query = `
			INSERT INTO attribute_values (attribute_id, value)
			VALUES ($1, $2)
			RETURNING id, attribute_id, value, is_active, created_at
		`
		err = tx.QueryRow(ctx, query, created.ID, value.Value).Scan(
			&val.ID, &val.AttributeID, &val.Value, &val.IsActive, &val.CreatedAt)
		if err != nil {
			fmt.Println("insertion into attribute_values table failed: ", err.Error())
			return nil, err
		}
		created.Values = append(created.Values, val)
	}

	err = tx.Commit(ctx)
	if err != nil {
		fmt.Println("commit failed: ", err.Error())
		return nil, err
	}
	return &created, nil
}

func (repo *ProductRepository) AddAttributeValues(ctx context.Context, attributeValues *domain.AddAttributeValuesRequest) error {

	query := `
		INSERT INTO attribute_values (attribute_id, value)
		VALUES ($1, $2)
	`

	for i := range attributeValues.Values {

		cmdTag, err := repo.DB.Exec(ctx, query, attributeValues.AttributeID, attributeValues.Values[i])
		if err != nil {
			return err
		}
		if cmdTag.RowsAffected() == 0 {
			return fmt.Errorf("ATTRIBUTE_NOT_FOUND")
		}
	}
	return nil
}

func (repo *ProductRepository) GetAttributes(ctx context.Context, activeOnly bool) ([]*domain.Attribute, error) {
	query := `
	SELECT 
		a.id,
		a.name,
		a.data_type,
		COALESCE(
			json_agg(
				json_build_object(
					'id', av.id,
					'value', av.value
				)
			) FILTER (WHERE av.id IS NOT NULL),
			'[]'
		) AS values
	FROM attributes a
	LEFT JOIN attribute_values av ON a.id = av.attribute_id`
	if activeOnly {
		query += ` WHERE a.is_active = true`
	}
	query += ` GROUP BY a.id, a.name, a.data_type ORDER BY a.id;`

	rows, err := repo.DB.Query(ctx, query)

	if err != nil {
		return nil, err
	}
	defer rows.Close()
	attributes := []*domain.Attribute{}
	for rows.Next() {
		var a domain.Attribute
		err := rows.Scan(&a.ID, &a.Name, &a.DataType, &a.Values)
		if err != nil {
			return nil, err
		}
		attributes = append(attributes, &a)
	}
	return attributes, nil
}

func (repo *ProductRepository) GetAttributeByID(ctx context.Context, id int64, activeOnly bool) (*domain.Attribute, error) {
	query := `
	SELECT 
		a.id,
		a.name,
		a.data_type,
		COALESCE(
			json_agg(
				json_build_object(
					'id', av.id,
					'value', av.value
				)
			) FILTER (WHERE av.id IS NOT NULL),
			'[]'
		) AS values
	FROM attributes a
	LEFT JOIN attribute_values av ON a.id = av.attribute_id WHERE a.id = $1`
	if activeOnly {
		query += ` AND a.is_active = true`
	}
	query += " GROUP BY a.id, a.name ORDER BY a.id;"
	fmt.Println("query : ", query, "id : ", id)
	var attribute domain.Attribute
	err := repo.DB.QueryRow(ctx, query, id).Scan(&attribute.ID, &attribute.Name, &attribute.DataType, &attribute.Values)
	if err != nil {
		return nil, err
	}
	return &attribute, nil
}

func (repo *ProductRepository) DeleteAttribute(ctx context.Context, id int64) error {
	query := `
	DELETE FROM attributes WHERE id = $1
	`
	cmdTag, err := repo.DB.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("ATTRIBUTE_NOT_FOUND")
	}
	return nil
}

func (repo *ProductRepository) DeleteAttributeValues(ctx context.Context, req *domain.DeleteAttributeValuesRequest) error {
	query := `
		DELETE FROM attribute_values WHERE id = $1
	`
	for i := range req.ValueIDs {
		cmdTag, err := repo.DB.Exec(ctx, query, req.ValueIDs[i])
		if err != nil {
			return err
		}
		if cmdTag.RowsAffected() == 0 {
			return fmt.Errorf("ATTRIBUTE_VALUE_NOT_FOUND")
		}
	}
	return nil
}
