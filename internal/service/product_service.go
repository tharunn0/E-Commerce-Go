package service

import (
	"context"

	"github.com/tharunn0/E-Commerce-Go/internal/domain"
	"github.com/tharunn0/E-Commerce-Go/internal/infrastructure/repository"
	"github.com/tharunn0/E-Commerce-Go/internal/utils"
	"go.uber.org/zap"
)

type ProductService struct {
	repo repository.ProductRepository
	log  *zap.Logger
}

func NewProductService(repo repository.ProductRepository, log *zap.Logger) *ProductService {
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

func (serv *ProductService) GetAllBrands(ctx context.Context) ([]domain.Brand, *domain.APIError) {
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
