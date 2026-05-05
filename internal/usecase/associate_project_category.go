package usecase

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// AssociateProjectCategory persists a project-category-agent link and copies
// every skill currently in the category into the project's agent skills folder.
// Skills already copied are silently skipped.
type AssociateProjectCategory struct {
	categories CategoryRepository
	projects   ProjectRepository
}

func NewAssociateProjectCategory(categories CategoryRepository, projects ProjectRepository) *AssociateProjectCategory {
	return &AssociateProjectCategory{categories: categories, projects: projects}
}

type AssociateProjectCategoryRequest struct {
	ProjectID  string
	CategoryID int64
	Agent      string
}

func (uc *AssociateProjectCategory) Execute(ctx context.Context, req AssociateProjectCategoryRequest) error {
	project, err := uc.projects.GetByID(ctx, req.ProjectID)
	if err != nil {
		return fmt.Errorf("associate project category: get project: %w", err)
	}

	if err := uc.categories.LinkProject(ctx, req.ProjectID, req.CategoryID, req.Agent); err != nil {
		return fmt.Errorf("associate project category: link: %w", err)
	}

	skillPaths, err := uc.categories.ListCategorySkillPaths(ctx, req.CategoryID)
	if err != nil {
		return fmt.Errorf("associate project category: list skill paths: %w", err)
	}

	dstParent := agentSkillsDir(project.Path, req.Agent)
	_ = os.MkdirAll(dstParent, 0o755)

	for _, sp := range skillPaths {
		dstDir := filepath.Join(dstParent, filepath.Base(sp.Path))
		if _, err := os.Stat(dstDir); err == nil {
			continue // already exists
		}
		_ = copyDir(sp.Path, dstDir) // best-effort
	}

	return nil
}

// DisassociateProjectCategory removes the project-category-agent link.
// Copied skill directories are intentionally NOT removed to avoid data loss.
type DisassociateProjectCategory struct {
	categories CategoryRepository
}

func NewDisassociateProjectCategory(categories CategoryRepository) *DisassociateProjectCategory {
	return &DisassociateProjectCategory{categories: categories}
}

type DisassociateProjectCategoryRequest struct {
	ProjectID  string
	CategoryID int64
	Agent      string
}

func (uc *DisassociateProjectCategory) Execute(ctx context.Context, req DisassociateProjectCategoryRequest) error {
	if err := uc.categories.UnlinkProject(ctx, req.ProjectID, req.CategoryID, req.Agent); err != nil {
		return fmt.Errorf("disassociate project category: %w", err)
	}
	return nil
}
