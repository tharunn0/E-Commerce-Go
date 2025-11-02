package service

import (
	"context"

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
func (serv *ProductService) CreateBrand(ctx context.Context, createBrandRequest *domain.CreateBrandRequest) (*domain.Brand, *domain.APIError) {

	brand, err := serv.repo.CreateBrand(ctx, createBrandRequest)
	if err != nil {
		serv.log.Debug("failed to create brand", zap.Error(err))
		return nil, &domain.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to create brand",
		}
	}
	return brand, nil
}

func (serv *ProductService) GetAllBrands(ctx context.Context) ([]*domain.Brand, *domain.APIError) {
	activeOnly := !utils.IsAdmin(ctx)
	brands, err := serv.repo.GetAllBrands(ctx, activeOnly)
	if err != nil {
		serv.log.Debug("failed to get all brands", zap.Error(err))
		return nil, &domain.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to get all brands",
		}
	}
	return brands, nil
}

func (serv *ProductService) GetBrandByID(ctx context.Context, id int64) (*domain.Brand, *domain.APIError) {

	activeOnly := !utils.IsAdmin(ctx)

	brand, err := serv.repo.GetBrandByID(ctx, id, activeOnly)
	if err != nil {
		serv.log.Debug("failed to get brand", zap.Error(err))
		return nil, &domain.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to get brand",
		}
	}
	return brand, nil
}

func (serv *ProductService) UpdateBrand(ctx context.Context, updateBrandRequest *domain.UpdateBrandRequest) (*domain.Brand, *domain.APIError) {

	brand, err := serv.repo.GetBrandByID(ctx, updateBrandRequest.ID, false)
	if err != nil {
		serv.log.Debug("failed to get brand", zap.Error(err))
		return nil, &domain.APIError{
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
		return nil, &domain.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to update brand",
		}
	}
	return updatedBrand, nil
}

func (serv *ProductService) DeleteBrand(ctx context.Context, id int64) *domain.APIError {
	err := serv.repo.DeleteBrand(ctx, id)
	if err != nil {
		serv.log.Debug("failed to delete brand", zap.Error(err))
		return &domain.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to delete brand",
		}
	}
	return nil
}

// Product operations
func (serv *ProductService) CreateProduct(ctx context.Context, createProductRequest *domain.CreateProductRequest) (*domain.Product, *domain.APIError) {
	createdProduct, err := serv.repo.CreateProduct(ctx, createProductRequest)
	if err != nil {
		serv.log.Debug("failed to create product", zap.Error(err))
		return nil, &domain.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to create product",
		}
	}
	return createdProduct, nil
}

func (serv *ProductService) GetProducts(ctx context.Context, page, limit int) ([]*domain.ProductResponse, int64, *domain.APIError) {

	activeOnly := !utils.IsAdmin(ctx)
	products, total, err := serv.repo.GetProducts(ctx, int64(page), int64(limit), activeOnly)
	if err != nil {
		serv.log.Debug("failed to get products", zap.String("function", "GetProducts"), zap.Error(err))
		return nil, 0, &domain.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to get products",
		}
	}
	return products, total, nil
}

func (serv *ProductService) GetProductByID(ctx context.Context, id int64) (*domain.ProductResponse, *domain.APIError) {
	activeOnly := !utils.IsAdmin(ctx)
	product, err := serv.repo.GetProductByID(ctx, id, activeOnly)
	if err != nil {
		serv.log.Debug("failed to get product", zap.String("function", "GetProductByID"), zap.Error(err))
		return nil, &domain.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to get product",
		}
	}
	return product, nil
}

func (serv *ProductService) UpdateProduct(ctx context.Context, updateProductRequest *domain.UpdateProductRequest) (*domain.ProductResponse, *domain.APIError) {

	err := serv.repo.UpdateProduct(ctx, updateProductRequest)
	if err != nil {
		serv.log.Debug("failed to update product", zap.Error(err))
		return nil, &domain.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to update product",
		}
	}

	updatedProduct, err := serv.repo.GetProductByID(ctx, updateProductRequest.ID, false)
	if err != nil {
		serv.log.Debug("failed to get product", zap.Error(err))
		return nil, &domain.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to get product",
		}
	}

	return updatedProduct, nil

}

func (serv *ProductService) DeleteProduct(ctx context.Context, id int64) *domain.APIError {
	err := serv.repo.DeleteProduct(ctx, id)
	if err != nil {
		serv.log.Debug("failed to delete product", zap.Error(err))
		return &domain.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to delete product",
		}
	}
	return nil
}
