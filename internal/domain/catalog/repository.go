package catalog

import (
	"context"
)

type CategoryRepository interface {
	CreateCategory(ctx context.Context, category *Category) (*Category, error)
	UpdateCategory(ctx context.Context, category *Category) (*Category, error)
	ChangeCategoryStatus(ctx context.Context, id int64, isActive bool) (*Category, error)
	GetCategoryByID(ctx context.Context, id int64, activeOnly bool) (*Category, error)
	GetAllCategories(ctx context.Context, activeOnly bool) ([]*Category, error)
	DeleteCategory(ctx context.Context, id int64) error
}
