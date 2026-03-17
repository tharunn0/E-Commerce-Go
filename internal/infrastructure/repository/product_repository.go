package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tharunn0/E-Commerce-Go/internal/apperror"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/product"
)

type ProductRepository struct {
	DB *pgxpool.Pool
}

func NewProductRepository(db *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{DB: db}
}

// Brand operations
// //////////////////////////////////////////////////////////
func (repo *ProductRepository) CreateBrand(ctx context.Context, createBrandRequest *product.CreateBrandRequest) (*product.Brand, error) {
	query := `
		INSERT INTO brands (name, description, logo_url)
		VALUES ($1, $2, $3)
		RETURNING id, name, description, logo_url, is_active, created_at, updated_at
	`
	var created product.Brand
	err := repo.DB.QueryRow(ctx, query, createBrandRequest.Name, createBrandRequest.Description, createBrandRequest.LogoURL).Scan(
		&created.ID, &created.Name, &created.Description, &created.LogoURL, &created.IsActive, &created.CreatedAt, &created.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &created, nil
}

func (repo *ProductRepository) GetAllBrands(ctx context.Context, activeOnly bool) ([]*product.Brand, error) {

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

	brands := []*product.Brand{}
	for rows.Next() {
		var b product.Brand
		err := rows.Scan(&b.ID, &b.Name, &b.Description, &b.LogoURL, &b.IsActive, &b.CreatedAt, &b.UpdatedAt)
		if err != nil {
			return nil, err
		}
		brands = append(brands, &b)
	}

	return brands, nil
}

func (repo *ProductRepository) GetBrandByID(ctx context.Context, id int64, activeOnly bool) (*product.Brand, error) {
	query := `
		SELECT id, name, description, logo_url, is_active, created_at, updated_at
		FROM brands
		WHERE id = $1
	`
	if activeOnly {
		query += " AND is_active = true"
	}
	var brand product.Brand
	err := repo.DB.QueryRow(ctx, query, id).Scan(&brand.ID, &brand.Name, &brand.Description, &brand.LogoURL, &brand.IsActive, &brand.CreatedAt, &brand.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &brand, nil
}

func (repo *ProductRepository) UpdateBrand(ctx context.Context, brand *product.Brand) (*product.Brand, error) {
	query := `
		UPDATE brands SET name = $1, description = $2, logo_url = $3, is_active = $4, updated_at = now() WHERE id = $5
		RETURNING id, name, description, logo_url, is_active, created_at, updated_at
	`
	var updated product.Brand
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
func (repo *ProductRepository) CreateProduct(ctx context.Context, productReq *product.CreateProductRequest) (*product.Product, error) {

	if productReq.MinPrice == nil {
		*productReq.MinPrice = 0
	}
	if productReq.MaxPrice == nil {
		*productReq.MaxPrice = 0
	}

	query := `
		INSERT INTO products (name, brand_id, description, category_id, min_price, max_price, is_digital, is_active, image_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, name, brand_id, description, category_id, min_price, max_price, is_digital, is_active, image_url, created_at, updated_at
	`
	var created product.Product
	err := repo.DB.QueryRow(ctx, query, productReq.Name, productReq.BrandID, productReq.Description, productReq.CategoryId,
		productReq.MinPrice, productReq.MaxPrice, productReq.IsDigital, productReq.IsActive, productReq.ImageURL).Scan(&created.ID, &created.Name, &created.BrandID, &created.Description,
		&created.CategoryId, &created.MinPrice, &created.MaxPrice, &created.IsDigital, &created.IsActive, &created.ImageURL, &created.CreatedAt, &created.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &created, nil
}

func (repo *ProductRepository) GetProducts(ctx context.Context, filter *product.ProductFilter, activeOnly bool) ([]*product.ProductResponse, int64, error) {

	var total int64
	query := `
		SELECT COUNT(*) FROM products
	`
	err := repo.DB.QueryRow(ctx, query).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query = `
		SELECT p.id, p.name,b.id as brand_id, b.name as brand_name, COALESCE(p.description, ''),
		 c.id as category_id, c.name as category_name,COALESCE(r.avg_rating,0) as rating, p.min_price, p.max_price, p.is_digital, p.is_active, p.image_url, p.created_at, p.updated_at
		FROM products p
		LEFT JOIN brands b ON p.brand_id = b.id
		LEFT JOIN categories c ON p.category_id = c.id
		LEFT JOIN (
			SELECT product_id,AVG(rating) as avg_rating FROM reviews
			GROUP BY product_id	
		) r ON r.product_id = p.id
		WHERE 1=1
	`
	if activeOnly {
		query += " AND p.is_active = true"
	}
	args := []any{}
	argIndex := 1

	if filter.Search != "" {
		query += fmt.Sprintf(" AND (p.name ILIKE $%d OR b.name ILIKE $%d OR c.name ILIKE $%d)", argIndex, argIndex, argIndex)
		args = append(args, "%"+filter.Search+"%")
		argIndex++
	}

	if filter.BrandID != nil && *filter.BrandID >= 0 {
		query += fmt.Sprintf(" AND p.brand_id = %d", *filter.BrandID)
	}

	if filter.CategoryID != nil && *filter.CategoryID >= 0 {
		query += fmt.Sprintf(" AND p.category_id = %d", *filter.CategoryID)
	}

	if filter.MinPrice > 0 {
		query += fmt.Sprintf(" AND p.min_price > %.2f", filter.MinPrice)
	}
	if filter.MaxPrice > 0 {
		query += fmt.Sprintf(" AND p.max_price < %.2f", filter.MaxPrice)
	}

	if filter.Sort != "" {
		if filter.Sort == "rating" {
			query += fmt.Sprintf(" ORDER BY %s %s", filter.Sort, filter.Order)
		} else {
			query += fmt.Sprintf(" ORDER BY p.%s %s", filter.Sort, filter.Order)
		}
	}
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filter.Limit)
		argIndex++
	}
	offset := (filter.Page - 1) * filter.Limit
	if filter.Page > 0 {
		query += fmt.Sprintf(" OFFSET $%d\n", argIndex)
		args = append(args, offset)
		argIndex++
	}

	log.Println("query", query)

	rows, err := repo.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	products := []*product.ProductResponse{}
	for rows.Next() {
		var p product.ProductResponse
		var b product.ProductBrandResponse
		var c product.ProductCategoryResponse
		if err := rows.Scan(&p.ID, &p.Name, &b.ID, &b.Name, &p.Description, &c.ID, &c.Name,
			&p.Rating, &p.MinPrice, &p.MaxPrice, &p.IsDigital, &p.IsActive, &p.ImageURL, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, 0, err
		}
		p.Brand = &b
		p.Category = &c
		products = append(products, &p)
	}

	total = int64(len(products))

	return products, total, nil
}

func (repo *ProductRepository) GetProductByID(ctx context.Context, id int64, activeOnly bool) (*product.ProductResponse, error) {
	query := `
		SELECT p.id, p.name, p.brand_id,b.name as brand_name, COALESCE(p.description, ''), p.category_id,c.name as category_name,
	 	p.min_price, p.max_price, p.is_digital, p.is_active, p.image_url, p.created_at, p.updated_at FROM products p
		LEFT JOIN brands b ON p.brand_id = b.id
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.id = $1
	`
	if activeOnly {
		query += " AND p.is_active = true"
	}
	var p product.ProductResponse
	var b product.ProductBrandResponse
	var c product.ProductCategoryResponse
	err := repo.DB.QueryRow(ctx, query, id).Scan(&p.ID, &p.Name, &b.ID, &b.Name, &p.Description, &c.ID, &c.Name,
		&p.MinPrice, &p.MaxPrice, &p.IsDigital, &p.IsActive, &p.ImageURL, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperror.ErrProductDoesNotExist
		}
		log.Println("Failed to get product by id", "[id : ", id, "] [error : ", err, "]")
		return nil, err
	}
	p.Brand = &b
	p.Category = &c

	return &p, nil
}

func (repo *ProductRepository) UpdateProduct(ctx context.Context, product *product.UpdateProductRequest) error {
	query := `UPDATE products
	SET name = COALESCE($1,name), brand_id = COALESCE($2,brand_id), description = COALESCE($3,description),
	 category_id = COALESCE($4,category_id), image_url = COALESCE($5,image_url), min_price = COALESCE($6,min_price), max_price = COALESCE($7,max_price),
	 is_digital = COALESCE($8,is_digital), is_active = COALESCE($9,is_active), updated_at = now()
	WHERE id = $9
	`
	cmdTag, err := repo.DB.Exec(ctx, query, product.Name, product.BrandID, product.Description,
		product.CategoryId, product.ImageURL, product.MinPrice, product.MaxPrice, product.IsDigital, product.IsActive, product.ID)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("PRODUCT_NOT_FOUND")
	}
	return nil
}

func (repo *ProductRepository) ToggleProductStatus(ctx context.Context, req *product.ProductStatusRequest) error {

	query := `UPDATE products SET is_active = $1 WHERE id = $2`

	cmdTag, err := repo.DB.Exec(ctx, query, req.Status, req.ID)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("Failed to update status. Product not found")
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
func (repo *ProductRepository) CreateProductVariant(ctx context.Context, req *product.CreateProductVariantRequest) (*product.ProductVariantResponse, error) {

	var created product.ProductVariantResponse
	// start transaction
	tx, err := repo.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// insert product variant
	query := `
		INSERT INTO product_variants (product_id, sku, original_price, sale_price, stock)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, product_id, sku, original_price, sale_price, stock, created_at`
	err = tx.QueryRow(ctx, query, req.ProductID, req.SKU, req.OriginalPrice, req.SalePrice, req.Stock).Scan(
		&created.ID, &created.BaseProduct.ID, &created.SKU, &created.OriginalPrice, &created.SalePrice, &created.Stock, &created.CreatedAt)
	if err != nil {
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
		var attrVal product.AttributeValueResponse

		err = tx.QueryRow(ctx,
			query, created.ID,
			req.VariantAttributes[i].AttributeID,
			req.VariantAttributes[i].AttributeValueID).Scan(
			&attrVal.ID)
		if err != nil {
			return nil, err
		}
		created.Attributes = append(created.Attributes, attrVal)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	// get base product
	query = `SELECT p.id,p.name,b.id,b.name FROM products p
	LEFT JOIN brands b ON p.brand_id = b.id
	WHERE p.id = $1
	`
	err = repo.DB.QueryRow(ctx, query, req.ProductID).Scan(
		&created.BaseProduct.ID,
		&created.BaseProduct.Name,
		&created.BaseProduct.Brand.ID,
		&created.BaseProduct.Brand.Name)
	if err != nil {
		return nil, err
	}

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
			return nil, err
		}
	}
	return &created, nil
}

func (repo *ProductRepository) GetProductVariantByID(ctx context.Context, id int64, activeOnly bool) (*product.ProductVariantResponse, error) {

	variant := product.ProductVariantResponse{}
	baseProduct := &product.BaseProduct{}
	variant.BaseProduct = baseProduct
	// get variant
	query := `
		SELECT id, product_id, sku, original_price, sale_price, stock, created_at FROM product_variants WHERE id = $1
		`
	if activeOnly {
		query += " AND is_active = true"
	}
	err := repo.DB.QueryRow(ctx, query, id).Scan(&variant.ID, &variant.BaseProduct.ID, &variant.SKU, &variant.OriginalPrice, &variant.SalePrice, &variant.Stock, &variant.CreatedAt)
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
		variant.VariantImages = append(variant.VariantImages, image)
	}

	// get variant attributes
	query = `
	SELECT pva.id, a.name,av.value FROM product_variant_attributes pva
	LEFT JOIN attributes a ON pva.attribute_id = a.id
	LEFT JOIN attribute_values av ON pva.attribute_value_id = av.id
	WHERE pva.product_variant_id = $1
	`
	rows, err = repo.DB.Query(ctx, query, variant.ID)
	if err != nil {
		return nil, err
	}
	var attrVal product.AttributeValueResponse
	for rows.Next() {
		err = rows.Scan(&attrVal.ID, &attrVal.Attribute, &attrVal.Value)
		if err != nil {
			return nil, err
		}
		variant.Attributes = append(variant.Attributes, attrVal)
	}
	// get base product
	query = `SELECT p.name, br.id, br.name,c.id FROM products p
			LEFT JOIN brands br ON br.id = p.brand_id
			LEFT JOIN categories c ON c.id = p.category_id
			WHERE p.id = $1`
	err = repo.DB.QueryRow(ctx, query, variant.BaseProduct.ID).Scan(
		&variant.BaseProduct.Name, &variant.BaseProduct.Brand.ID, &variant.BaseProduct.Brand.Name, &variant.BaseProduct.CategoryID)
	if err != nil {
		return nil, err
	}

	// update min and max price

	return &variant, nil
}

func (repo *ProductRepository) GetVariantsByProductID(ctx context.Context, productID int64, activeOnly bool) (*product.ProductVariantBaseResponse, error) {
	productvariants := product.ProductVariantBaseResponse{}

	query := `
		SELECT p.id, p.name as product_name, b.name as brand_name,p.min_price, p.max_price, p.is_digital,p.image_url, p.created_at FROM products p
		 LEFT JOIN brands b ON p.brand_id = b.id WHERE p.id = $1 AND p.is_active = true
	`
	err := repo.DB.QueryRow(ctx, query, productID).Scan(
		&productvariants.ProductID,
		&productvariants.ProductName,
		&productvariants.BrandName,
		&productvariants.MinPrice,
		&productvariants.MaxPrice,
		&productvariants.IsDigital,
		&productvariants.ImageURL,
		&productvariants.CreatedAt)
	if err != nil {
		return nil, err
	}

	var baseProduct product.BaseProduct

	query = `SELECT p.name, br.id, br.name,c.id FROM products p
			LEFT JOIN brands br ON br.id = p.brand_id
			LEFT JOIN categories c ON c.id = p.category_id
			WHERE p.id = $1`
	err = repo.DB.QueryRow(ctx, query, productID).Scan(
		&baseProduct.Name, &baseProduct.Brand.ID, &baseProduct.Brand.Name, &baseProduct.CategoryID)
	if err != nil {
		return nil, err
	}

	// get variants
	query = `
		SELECT id, sku, original_price, sale_price, stock, is_active, created_at FROM product_variants WHERE product_id = $1 AND is_active = true
	`
	rows, err := repo.DB.Query(ctx, query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var variant product.ProductVariantResponse
		err := rows.Scan(&variant.ID, &variant.SKU, &variant.OriginalPrice, &variant.SalePrice, &variant.Stock, &variant.IsActive, &variant.CreatedAt)
		if err != nil {
			return nil, err
		}
		productvariants.Variants = append(productvariants.Variants, &variant)
	}

	for i := range productvariants.Variants {
		query = `
		SELECT  url FROM product_variant_images WHERE product_variant_id = $1
	`
		rows, err = repo.DB.Query(ctx, query, productvariants.Variants[i].ID)
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
			productvariants.Variants[i].VariantImages = append(productvariants.Variants[i].VariantImages, image)
		}

		if productvariants.Variants[i].SalePrice != nil && *productvariants.Variants[i].SalePrice == 0 {
			productvariants.Variants[i].SalePrice = nil
		}

		// set base product
		productvariants.Variants[i].BaseProduct = &baseProduct

	}

	// get variant attributes
	for i := range productvariants.Variants {
		query = `
		SELECT pva.id,a.name as attribute,av.value as value FROM product_variant_attributes pva
		LEFT JOIN attributes a ON pva.attribute_id = a.id
		LEFT JOIN attribute_values av ON pva.attribute_value_id = av.id
		WHERE pva.product_variant_id = $1
	`
		rows, err = repo.DB.Query(ctx, query, productvariants.Variants[i].ID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var attrVal product.AttributeValueResponse
			err := rows.Scan(&attrVal.ID, &attrVal.Attribute, &attrVal.Value)
			if err != nil {
				return nil, err
			}
			productvariants.Variants[i].Attributes = append(productvariants.Variants[i].Attributes, attrVal)
		}
	}
	return &productvariants, nil
}

func (repo *ProductRepository) UpdateProductVariant(ctx context.Context, productVariant *product.UpdateProductVariantRequest) (*product.ProductVariant, error) {
	query := `UPDATE product_variants SET sku = COALESCE($1,sku), original_price = COALESCE($2,original_price), sale_price = COALESCE($3,sale_price),
	 stock = COALESCE($4,stock), is_active = COALESCE($5,is_active), updated_at = now() WHERE id = $6
	RETURNING id, product_id, sku, original_price, sale_price, stock, is_active, created_at
	`
	var updated product.ProductVariant
	err := repo.DB.QueryRow(ctx, query,
		productVariant.SKU,
		productVariant.OriginalPrice,
		productVariant.SalePrice,
		productVariant.Stock,
		productVariant.IsActive,
		productVariant.ID).Scan(
		&updated.ID, &updated.ProductID, &updated.SKU, &updated.OriginalPrice, &updated.SalePrice, &updated.Stock, &updated.IsActive, &updated.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

func (repo *ProductRepository) ToggleVariantStatus(ctx context.Context, req *product.VariantStatusRequest) error {
	query := `UPDATE product_variants SET is_active = $1 WHERE id = $2`
	cmdTag, err := repo.DB.Exec(ctx, query, req.Status, req.ID)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("Failed to update status. Product variant not found")
	}
	return nil
}

func (repo *ProductRepository) DeleteProductVariant(ctx context.Context, id int64) error {
	cmdTag, err := repo.DB.Exec(ctx, `DELETE FROM product_variants WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("PRODUCT_VARIANT_NOT_FOUND")
	}
	return nil
}

func (repo *ProductRepository) UpdateProductMinMaxPrice(ctx context.Context, productID int64) error {
	query := `
		UPDATE products SET min_price = (SELECT MIN(original_price) FROM product_variants WHERE product_id = $1),
		max_price = (SELECT MAX(original_price) FROM product_variants WHERE product_id = $1) WHERE id = $1
	`
	cmdTag, err := repo.DB.Exec(ctx, query, productID)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("PRODUCT_NOT_FOUND")
	}
	return nil
}

func (repo *ProductRepository) GetProductVariantOrderInfo(ctx context.Context, productVariantID int64) (*product.VariantOrderInfo, error) {
	query := `
		SELECT p.name, pv.id as product_variant_id, pv.sku
		FROM products p
		JOIN product_variants pv ON p.id = pv.product_id
		WHERE pv.id = $1
	`
	var info product.VariantOrderInfo
	err := repo.DB.QueryRow(ctx, query, productVariantID).Scan(&info.ProductName, &info.ProductVariantID, &info.SKU)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

// Attribute operations
// //////////////////////////////////////////////////////////
func (repo *ProductRepository) CreateAttribute(ctx context.Context, attribute *product.CreateAttributeRequest) (*product.Attribute, error) {

	var created product.Attribute
	tx, err := repo.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	query := `
		INSERT INTO attributes (name, data_type)
		VALUES ($1, $2)
		RETURNING id, name, data_type, is_active
	`
	err = tx.QueryRow(ctx, query, attribute.Name, attribute.DataType).Scan(
		&created.ID, &created.Name, &created.DataType, &created.IsActive)
	if err != nil {
		return nil, err
	}

	for _, value := range attribute.Values {
		var val product.AttributeValue

		query = `
			INSERT INTO attribute_values (attribute_id, value)
			VALUES ($1, $2)
			RETURNING id, attribute_id, value, is_active
		`
		err = tx.QueryRow(ctx, query, created.ID, value.Value).Scan(
			&val.ID, &val.AttributeID, &val.Value, &val.IsActive)
		if err != nil {
			return nil, err
		}
		created.Values = append(created.Values, val)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}
	return &created, nil
}

func (repo *ProductRepository) AddAttributeValues(ctx context.Context, attributeValues *product.AddAttributeValuesRequest) error {

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

func (repo *ProductRepository) GetAttributes(ctx context.Context, activeOnly bool) ([]*product.Attribute, error) {
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
	attributes := []*product.Attribute{}
	for rows.Next() {
		var a product.Attribute
		err := rows.Scan(&a.ID, &a.Name, &a.DataType, &a.Values)
		if err != nil {
			return nil, err
		}
		attributes = append(attributes, &a)
	}
	return attributes, nil
}

func (repo *ProductRepository) GetAttributeByID(ctx context.Context, id int64, activeOnly bool) (*product.Attribute, error) {
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
	var attribute product.Attribute
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

func (repo *ProductRepository) DeleteAttributeValues(ctx context.Context, req *product.DeleteAttributeValuesRequest) error {
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
