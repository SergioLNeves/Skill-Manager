package usecase

import (
	"context"
	"fmt"

	"skill-manager/internal/domain"
)

// ManageCategories handles CRUD operations for categories.
type ManageCategories struct {
	categories CategoryRepository
}

func NewManageCategories(categories CategoryRepository) *ManageCategories {
	return &ManageCategories{categories: categories}
}

func (uc *ManageCategories) Create(ctx context.Context, name, description string) (domain.Category, error) {
	if name == "" {
		return domain.Category{}, fmt.Errorf("manage categories: name is required")
	}
	return uc.categories.Create(ctx, name, description)
}

func (uc *ManageCategories) Update(ctx context.Context, id int64, name, description string) error {
	if name == "" {
		return fmt.Errorf("manage categories: name is required")
	}
	return uc.categories.Update(ctx, id, name, description)
}

func (uc *ManageCategories) Delete(ctx context.Context, id int64) error {
	return uc.categories.Delete(ctx, id)
}

func (uc *ManageCategories) List(ctx context.Context) ([]domain.Category, error) {
	return uc.categories.List(ctx)
}
