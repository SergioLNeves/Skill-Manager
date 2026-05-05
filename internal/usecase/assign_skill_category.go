package usecase

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// AssignSkillCategory sets (or clears) the category for a skill and propagates
// the change: for each project already linked to the new category (and agent),
// the skill is copied into that project if not already present.
type AssignSkillCategory struct {
	categories CategoryRepository
	projects   ProjectRepository
}

func NewAssignSkillCategory(categories CategoryRepository, projects ProjectRepository) *AssignSkillCategory {
	return &AssignSkillCategory{categories: categories, projects: projects}
}

type AssignSkillCategoryRequest struct {
	SkillName    string
	SkillPath    string  // filesystem path of the skill directory
	CategoryID   *int64  // nil to unassign
}

func (uc *AssignSkillCategory) Execute(ctx context.Context, req AssignSkillCategoryRequest) error {
	if err := uc.categories.AssignSkill(ctx, req.SkillName, req.CategoryID); err != nil {
		return fmt.Errorf("assign skill category: %w", err)
	}

	if req.CategoryID == nil || req.SkillPath == "" {
		return nil
	}

	// Propagate to projects already linked to this category.
	links, err := uc.categories.GetCategoryLinks(ctx, *req.CategoryID)
	if err != nil {
		return fmt.Errorf("assign skill category: get links: %w", err)
	}

	for _, link := range links {
		project, err := uc.projects.GetByID(ctx, link.ProjectID)
		if err != nil {
			continue
		}
		dstParent := agentSkillsDir(project.Path, string(link.Agent))
		dstDir := filepath.Join(dstParent, filepath.Base(req.SkillPath))
		if _, err := os.Stat(dstDir); err == nil {
			continue // already exists
		}
		if err := os.MkdirAll(dstParent, 0o755); err != nil {
			continue
		}
		_ = copyDir(req.SkillPath, dstDir) // best-effort
	}

	return nil
}
