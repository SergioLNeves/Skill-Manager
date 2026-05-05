package binding

import (
	"context"
	"fmt"

	"skill-manager/internal/usecase"
)

// CategoriesBinding exposes category-related operations to the Wails frontend.
type CategoriesBinding struct {
	manage       *usecase.ManageCategories
	assign       *usecase.AssignSkillCategory
	associate    *usecase.AssociateProjectCategory
	disassociate *usecase.DisassociateProjectCategory
	catRepo      usecase.CategoryRepository
}

func NewCategoriesBinding(
	manage *usecase.ManageCategories,
	assign *usecase.AssignSkillCategory,
	associate *usecase.AssociateProjectCategory,
	disassociate *usecase.DisassociateProjectCategory,
	catRepo usecase.CategoryRepository,
) *CategoriesBinding {
	return &CategoriesBinding{
		manage:       manage,
		assign:       assign,
		associate:    associate,
		disassociate: disassociate,
		catRepo:      catRepo,
	}
}

// ListCategories returns all categories ordered by name.
func (b *CategoriesBinding) ListCategories(ctx context.Context) ([]CategoryDTO, error) {
	cats, err := b.manage.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("categories binding: list: %w", err)
	}
	dtos := make([]CategoryDTO, len(cats))
	for i, c := range cats {
		dtos[i] = toCategoryDTO(c)
	}
	return dtos, nil
}

// CreateCategory creates a new category.
func (b *CategoriesBinding) CreateCategory(ctx context.Context, req CreateCategoryRequestDTO) (CategoryDTO, error) {
	cat, err := b.manage.Create(ctx, req.Name, req.Description)
	if err != nil {
		return CategoryDTO{}, fmt.Errorf("categories binding: create: %w", err)
	}
	return toCategoryDTO(cat), nil
}

// UpdateCategory updates name and description of an existing category.
func (b *CategoriesBinding) UpdateCategory(ctx context.Context, req UpdateCategoryRequestDTO) error {
	return b.manage.Update(ctx, req.ID, req.Name, req.Description)
}

// DeleteCategory deletes a category (cascades to project_categories links).
func (b *CategoriesBinding) DeleteCategory(ctx context.Context, id int64) error {
	return b.manage.Delete(ctx, id)
}

// AssignSkillCategory sets or clears the category for a skill.
func (b *CategoriesBinding) AssignSkillCategory(ctx context.Context, req AssignSkillCategoryRequestDTO) error {
	return b.assign.Execute(ctx, usecase.AssignSkillCategoryRequest{
		SkillName:  req.SkillName,
		SkillPath:  req.SkillPath,
		CategoryID: req.CategoryID,
	})
}

// AssociateProjectCategory persists a project-category-agent link and copies skills.
func (b *CategoriesBinding) AssociateProjectCategory(ctx context.Context, req ProjectCategoryRequestDTO) error {
	return b.associate.Execute(ctx, usecase.AssociateProjectCategoryRequest{
		ProjectID:  req.ProjectID,
		CategoryID: req.CategoryID,
		Agent:      req.Agent,
	})
}

// DisassociateProjectCategory removes the project-category-agent link (does not remove copied files).
func (b *CategoriesBinding) DisassociateProjectCategory(ctx context.Context, req ProjectCategoryRequestDTO) error {
	return b.disassociate.Execute(ctx, usecase.DisassociateProjectCategoryRequest{
		ProjectID:  req.ProjectID,
		CategoryID: req.CategoryID,
		Agent:      req.Agent,
	})
}

// ListProjectCategories returns all category links for the given project.
func (b *CategoriesBinding) ListProjectCategories(ctx context.Context, projectID string) ([]ProjectCategoryLinkDTO, error) {
	links, err := b.catRepo.ListProjectCategories(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("categories binding: list project categories: %w", err)
	}
	dtos := make([]ProjectCategoryLinkDTO, len(links))
	for i, l := range links {
		dtos[i] = toProjectCategoryLinkDTO(l)
	}
	return dtos, nil
}
