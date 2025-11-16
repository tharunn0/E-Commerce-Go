package service

import (
	"context"

	"github.com/tharunn0/E-Commerce-Go/internal/apperror"
	"github.com/tharunn0/E-Commerce-Go/internal/domain"
	"github.com/tharunn0/E-Commerce-Go/internal/utils"
	"go.uber.org/zap"
)

type ProductService struct {
	repo domain.ProductRepository
	log  *zap.Logger
}

func NewProductService(repo domain.ProductRepository, log *zap.Logger) *ProductService {
	return &ProductService{repo: repo, log: log}
}

// Brand operations
func (serv *ProductService) CreateBrand(ctx context.Context, createBrandRequest *domain.CreateBrandRequest) (*domain.Brand, *apperror.APIError) {

	brand, err := serv.repo.CreateBrand(ctx, createBrandRequest)
	if err != nil {
		serv.log.Debug("failed to create brand", zap.Error(err))
		return nil, &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to create brand",
		}
	}
	return brand, nil
}

func (serv *ProductService) GetAllBrands(ctx context.Context) ([]*domain.Brand, *apperror.APIError) {
	activeOnly := !utils.IsAdmin(ctx)
	brands, err := serv.repo.GetAllBrands(ctx, activeOnly)
	if err != nil {
		serv.log.Debug("failed to get all brands", zap.Error(err))
		return nil, &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to get all brands",
		}
	}
	return brands, nil
}

func (serv *ProductService) GetBrandByID(ctx context.Context, id int64) (*domain.Brand, *apperror.APIError) {

	activeOnly := !utils.IsAdmin(ctx)

	brand, err := serv.repo.GetBrandByID(ctx, id, activeOnly)
	if err != nil {
		serv.log.Debug("failed to get brand", zap.Error(err))
		return nil, &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to get brand",
		}
	}
	return brand, nil
}

func (serv *ProductService) UpdateBrand(ctx context.Context, updateBrandRequest *domain.UpdateBrandRequest) (*domain.Brand, *apperror.APIError) {

	brand, err := serv.repo.GetBrandByID(ctx, updateBrandRequest.ID, false)
	if err != nil {
		serv.log.Debug("failed to get brand", zap.Error(err))
		return nil, &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to get brand",
		}
	}

	if updateBrandRequest.Name != nil {
		brand.Name = *updateBrandRequest.Name
	}
	if updateBrandRequest.Description != nil {
		brand.Description = updateBrandRequest.Description
	}
	if updateBrandRequest.LogoURL != nil {
		brand.LogoURL = updateBrandRequest.LogoURL
	}
	if updateBrandRequest.IsActive != nil {
		brand.IsActive = *updateBrandRequest.IsActive
	}
	updatedBrand, err := serv.repo.UpdateBrand(ctx, brand)
	if err != nil {
		serv.log.Debug("failed to update brand", zap.Error(err))
		return nil, &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to update brand",
		}
	}
	return updatedBrand, nil
}

func (serv *ProductService) DeleteBrand(ctx context.Context, id int64) *apperror.APIError {
	err := serv.repo.DeleteBrand(ctx, id)
	if err != nil {
		serv.log.Debug("failed to delete brand", zap.Error(err))
		return &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to delete brand",
		}
	}
	return nil
}

// Product operations
func (serv *ProductService) CreateProduct(ctx context.Context, createProductRequest *domain.CreateProductRequest) (*domain.Product, *apperror.APIError) {
	createdProduct, err := serv.repo.CreateProduct(ctx, createProductRequest)
	if err != nil {
		serv.log.Debug("failed to create product", zap.Error(err))
		return nil, &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to create product",
		}
	}
	return createdProduct, nil
}

func (serv *ProductService) GetProducts(ctx context.Context, filter *domain.ProductFilter) ([]*domain.ProductResponse, int64, *apperror.APIError) {

	isAdmin := utils.IsAdmin(ctx)
	activeOnly := !utils.IsAdmin(ctx)

	if !utils.IsFiltersValid(filter) {
		return nil, 0, &apperror.APIError{
			Code:    "INVALID_FILTERS",
			Message: "Filters contain invalid values",
		}
	}

	products, total, err := serv.repo.GetProducts(ctx, filter, activeOnly)
	if err != nil {
		serv.log.Debug("failed to get products", zap.String("function", "GetProducts"), zap.Error(err))
		return nil, 0, &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to get products",
		}
	}

	if !isAdmin {
		for _, p := range products {
			p.CreatedAt = nil
			p.UpdatedAt = nil
			p.IsActive = nil
		}
	}

	return products, total, nil
}

func (serv *ProductService) GetProductByID(ctx context.Context, id int64) (*domain.ProductResponse, *apperror.APIError) {
	activeOnly := !utils.IsAdmin(ctx)
	product, err := serv.repo.GetProductByID(ctx, id, activeOnly)
	if err != nil {
		serv.log.Debug("failed to get product", zap.String("function", "GetProductByID"), zap.Error(err))
		return nil, &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to get product",
		}
	}
	if activeOnly {
		product.IsActive = nil
		product.CreatedAt = nil
		product.UpdatedAt = nil
	}
	return product, nil
}

func (serv *ProductService) UpdateProduct(ctx context.Context, updateProductRequest *domain.UpdateProductRequest) (*domain.ProductResponse, *apperror.APIError) {

	err := serv.repo.UpdateProduct(ctx, updateProductRequest)
	if err != nil {
		serv.log.Debug("failed to update product", zap.Error(err))
		return nil, &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to update product",
		}
	}

	updatedProduct, err := serv.repo.GetProductByID(ctx, updateProductRequest.ID, false)
	if err != nil {
		serv.log.Debug("failed to get product", zap.Error(err))
		return nil, &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to get product",
		}
	}

	return updatedProduct, nil
}

func (serv *ProductService) UpdateProductStatus(ctx context.Context, req *domain.ProductStatusRequest) *apperror.APIError {

	if req.ID <= 0 {
		return &apperror.APIError{
			Code:    "VALIDATION_ERROR",
			Message: "Invalid product ID provided",
		}
	}

	err := serv.repo.ToggleProductStatus(ctx, req)
	if err != nil {
		return &apperror.APIError{
			Code:    "DB_ERROR",
			Message: err.Error(),
		}
	}
	return nil
}

func (serv *ProductService) DeleteProduct(ctx context.Context, id int64) *apperror.APIError {
	err := serv.repo.DeleteProduct(ctx, id)
	if err != nil {
		serv.log.Debug("failed to delete product", zap.Error(err))
		return &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to delete product",
		}
	}
	return nil
}

// Product variant operations
func (serv *ProductService) CreateProductVariant(ctx context.Context, req *domain.CreateProductVariantRequest) (*domain.ProductVariantResponse, *apperror.APIError) {
	if len(req.VariantAttributes) == 0 {
		return nil, &apperror.APIError{
			Code:    "INVALID_REQUEST",
			Message: "Variant attributes are required",
		}
	}
	for _, variantAttribute := range req.VariantAttributes {
		if variantAttribute.AttributeID == nil || variantAttribute.AttributeValueID == nil {
			return nil, &apperror.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Attribute ID and value ID are required",
			}
		}
	}
	for _, image := range req.Images {
		if image == "" {
			return nil, &apperror.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Images are required",
			}
		}
	}
	if req.SKU == "" {
		return nil, &apperror.APIError{
			Code:    "INVALID_REQUEST",
			Message: "SKU is required",
		}
	}

	if req.OriginalPrice < 0 {
		return nil, &apperror.APIError{
			Code:    "INVALID_REQUEST",
			Message: "Original price cannot be negative",
		}
	}

	if *req.SalePrice == 0 {
		req.SalePrice = nil
	}

	productVariant, err := serv.repo.CreateProductVariant(ctx, req)
	if err != nil {
		serv.log.Error("failed to create product variant", zap.Error(err))
		return nil, &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to create product variant",
		}
	}
	err = serv.repo.UpdateProductMinMaxPrice(ctx, req.ProductID)
	if err != nil {
		serv.log.Error("failed to update product min max price", zap.Error(err))
		return nil, &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to update product min max price",
		}
	}
	return productVariant, nil
}

func (serv *ProductService) GetProductVariantByID(ctx context.Context, id int64) (*domain.ProductVariantResponse, *apperror.APIError) {

	activeOnly := !utils.IsAdmin(ctx)
	productVariant, err := serv.repo.GetProductVariantByID(ctx, id, activeOnly)
	if err != nil {
		serv.log.Debug("failed to get product variant", zap.Error(err))
		return nil, &apperror.APIError{
			Code:    "NOT_FOUND",
			Message: "Product variant not found",
		}
	}

	if *productVariant.SalePrice == 0 {
		productVariant.SalePrice = nil
	}

	return productVariant, nil
}

func (serv *ProductService) GetVariantsByProductID(ctx context.Context, productID int64) (*domain.ProductVariantBaseResponse, *apperror.APIError) {
	activeOnly := !utils.IsAdmin(ctx)
	productvariants, err := serv.repo.GetVariantsByProductID(ctx, productID, activeOnly)
	if err != nil {
		serv.log.Debug("failed to get variants by product ID", zap.Error(err))
		return nil, &apperror.APIError{
			Code:    "NOT_FOUND",
			Message: "Product variants not found",
		}
	}
	return productvariants, nil
}

func (serv *ProductService) UpdateProductVariant(ctx context.Context, updateProductVariantRequest *domain.UpdateProductVariantRequest) (*domain.ProductVariant, *apperror.APIError) {

	productVariant, err := serv.repo.UpdateProductVariant(ctx, updateProductVariantRequest)
	if err != nil {
		serv.log.Debug("failed to update product variant", zap.Error(err))
		return nil, &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to update product variant",
		}
	}

	err = serv.repo.UpdateProductMinMaxPrice(ctx, productVariant.ProductID)
	if err != nil {
		serv.log.Error("failed to update product min max price", zap.Error(err))
		return nil, &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to update product min max price",
		}
	}

	return productVariant, nil
}

func (serv *ProductService) UpdateProductVariantStatus(ctx context.Context, req *domain.VariantStatusRequest) *apperror.APIError {

	if req.ID <= 0 {
		return &apperror.APIError{
			Code:    "VALIDATION_ERROR",
			Message: "Invalid product variant ID provided",
		}
	}

	err := serv.repo.ToggleVariantStatus(ctx, req)
	if err != nil {
		return &apperror.APIError{
			Code:    "DB_ERROR",
			Message: err.Error(),
		}
	}
	return nil
}

func (serv *ProductService) DeleteProductVariant(ctx context.Context, id int64) *apperror.APIError {
	err := serv.repo.DeleteProductVariant(ctx, id)
	if err != nil {
		serv.log.Debug("failed to delete product variant", zap.Error(err))
		return &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to delete product variant",
		}
	}

	return nil
}

// Attribute operations
func (serv *ProductService) CreateAttribute(ctx context.Context, createAttributeRequest *domain.CreateAttributeRequest) (*domain.Attribute, *apperror.APIError) {
	if len(createAttributeRequest.Values) == 0 {
		return nil, &apperror.APIError{
			Code:    "INVALID_REQUEST",
			Message: "Values are required",
		}
	}

	if !utils.IsValidDataType(createAttributeRequest.DataType) {
		return nil, &apperror.APIError{
			Code:    "INVALID_REQUEST",
			Message: "Invalid data type",
		}
	}

	attribute, err := serv.repo.CreateAttribute(ctx, createAttributeRequest)
	if err != nil {
		serv.log.Debug("failed to create attribute", zap.Error(err))
		return nil, &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to create attribute",
		}
	}
	return attribute, nil
}

func (serv *ProductService) AddAttributeValues(ctx context.Context, addAttributeValuesRequest *domain.AddAttributeValuesRequest) *apperror.APIError {

	if len(addAttributeValuesRequest.Values) == 0 {
		return &apperror.APIError{
			Code:    "INVALID_REQUEST",
			Message: "Values are required",
		}
	}

	for _, value := range addAttributeValuesRequest.Values {
		if value == "" {
			return &apperror.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Values are required",
			}
		}
	}

	err := serv.repo.AddAttributeValues(ctx, addAttributeValuesRequest)
	if err != nil {
		serv.log.Debug("failed to add attribute values", zap.Error(err))
		return &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to add attribute values",
		}
	}
	return nil
}

func (serv *ProductService) GetAttributes(ctx context.Context) ([]*domain.Attribute, *apperror.APIError) {
	activeOnly := !utils.IsAdmin(ctx)
	attributes, err := serv.repo.GetAttributes(ctx, activeOnly)
	if err != nil {
		serv.log.Debug("failed to get attributes", zap.Error(err))
		return nil, &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to get attributes",
		}
	}
	return attributes, nil
}
func (serv *ProductService) GetAttributeByID(ctx context.Context, id int64) (*domain.Attribute, *apperror.APIError) {
	activeOnly := !utils.IsAdmin(ctx)
	attribute, err := serv.repo.GetAttributeByID(ctx, id, activeOnly)
	if err != nil {
		serv.log.Debug("failed to get attribute", zap.Error(err))
		return nil, &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to get attribute",
		}
	}
	return attribute, nil
}

func (serv *ProductService) DeleteAttribute(ctx context.Context, req *domain.DeleteAttributeRequest) *apperror.APIError {

	if req.ID <= 0 {
		return &apperror.APIError{
			Code:    "INVALID_DELETE_REQUEST",
			Message: "Provide a valid attribute id",
		}
	}

	err := serv.repo.DeleteAttribute(ctx, req.ID)
	if err != nil {
		serv.log.Debug("failed to delete attribute", zap.Error(err))
		return &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to delete attribute",
		}
	}
	return nil
}
func (serv *ProductService) DeleteAttributeValues(ctx context.Context, deleteAttributeValuesRequest *domain.DeleteAttributeValuesRequest) *apperror.APIError {
	if len(deleteAttributeValuesRequest.ValueIDs) == 0 {
		return &apperror.APIError{
			Code:    "INVALID_REQUEST",
			Message: "Value IDs are required",
		}
	}

	err := serv.repo.DeleteAttributeValues(ctx, deleteAttributeValuesRequest)
	if err != nil {
		serv.log.Debug("failed to delete attribute values", zap.Error(err))
		return &apperror.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to delete attribute values",
		}
	}
	return nil
}
