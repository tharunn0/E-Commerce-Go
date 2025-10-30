package service

import (
	"context"

	"github.com/tharunn0/E-Commerce-Go/internal/domain"
	"github.com/tharunn0/E-Commerce-Go/internal/infrastructure/repository"
	"github.com/tharunn0/E-Commerce-Go/internal/utils"
	"go.uber.org/zap"
)

type CategoryService struct {
	repo repository.CategoryRepository
	log  *zap.Logger
}

func NewCategoryService(repo repository.CategoryRepository, log *zap.Logger) *CategoryService {
	return &CategoryService{repo: repo, log: log}
}

func (s *CategoryService) CreateCategory(ctx context.Context, category *domain.Category) (*domain.Category, *domain.APIError) {

	category, err := s.repo.CreateCategory(ctx, category)
	if err != nil {
		s.log.Debug("failed to create category", zap.Error(err))
		return nil, &domain.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to create category",
		}
	}
	return category, nil

}

func (s *CategoryService) UpdateCategory(ctx context.Context, categoryreq *domain.UpdateCategoryRequest) (*domain.Category, *domain.APIError) {

	activeOnly := !utils.IsAdmin(ctx)
	category, err := s.repo.GetCategoryByID(ctx, categoryreq.ID, activeOnly)
	if err != nil {
		s.log.Debug("failed to get category by id", zap.Error(err))
		return nil, &domain.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to get category by id",
		}
	}

	if categoryreq.Name != nil {
		category.Name = *categoryreq.Name
	}
	if categoryreq.Description != nil {
		category.Description = *categoryreq.Description
	}
	if categoryreq.ParentCategoryID != nil {
		category.ParentCategoryID = categoryreq.ParentCategoryID
	}
	if categoryreq.ImageURL != nil {
		category.ImageURL = categoryreq.ImageURL
	}

	updatedCategory, err := s.repo.UpdateCategory(ctx, category)
	if err != nil {
		s.log.Debug("failed to update category", zap.Error(err))
		return nil, &domain.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to update category",
		}
	}
	return updatedCategory, nil
}

func (s *CategoryService) ChangeCategoryStatus(ctx context.Context, id int64, statusreq *domain.CategoryStatusRequest) (*domain.Category, *domain.APIError) {
	updatedCategory, err := s.repo.ChangeCategoryStatus(ctx, id, statusreq.IsActive)
	if err != nil {
		s.log.Debug("failed to change category status", zap.Error(err))
		return nil, &domain.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to change category status",
		}
	}
	return updatedCategory, nil
}

func (s *CategoryService) DeleteCategory(ctx context.Context, id int64) *domain.APIError {
	err := s.repo.DeleteCategory(ctx, id)
	if err != nil {
		s.log.Debug("failed to delete category", zap.Error(err))
		return &domain.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to delete category",
		}
	}
	return nil
}

func (s *CategoryService) GetCategoryByID(ctx context.Context, id int64) (*domain.Category, *domain.APIError) {
	activeOnly := utils.IsAdmin(ctx)
	category, err := s.repo.GetCategoryByID(ctx, id, activeOnly)
	if err != nil {
		s.log.Debug("failed to get category by id", zap.Error(err))
		return nil, &domain.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to get category by id",
		}
	}
	return category, nil
}

func (s *CategoryService) GetAllCategories(ctx context.Context) ([]*domain.Category, *domain.APIError) {
	activeOnly := !utils.IsAdmin(ctx)

	if activeOnly {
		s.log.Debug("its a user request", zap.Bool("activeOnly", activeOnly))
	} else {
		s.log.Debug("its an admin request", zap.Bool("activeOnly", activeOnly))
	}
	categories, err := s.repo.GetAllCategories(ctx, activeOnly)
	if err != nil {
		s.log.Debug("failed to get all categories", zap.Error(err))
		return nil, &domain.APIError{
			Code:    "DB_ERROR",
			Message: "Failed to get all categories",
		}
	}
	return categories, nil
}
