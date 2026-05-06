package usecase

import (
	"context"
	"fmt"
	"time"

	"skill-manager/internal/domain"
)

// AssociateProjectCategory persists a project-category-agent link, creates an
// Activation for every skill currently in the category, and syncs the adapter.
type AssociateProjectCategory struct {
	categories    CategoryRepository
	projects      ProjectRepository
	activations   ActivationRepository
	skills        SkillRepository
	projectSkills ProjectSkillRepository
	adapters      map[domain.Agent]AgentAdapter
}

func NewAssociateProjectCategory(
	categories CategoryRepository,
	projects ProjectRepository,
	activations ActivationRepository,
	skills SkillRepository,
	projectSkills ProjectSkillRepository,
	adapters map[domain.Agent]AgentAdapter,
) *AssociateProjectCategory {
	return &AssociateProjectCategory{
		categories:    categories,
		projects:      projects,
		activations:   activations,
		skills:        skills,
		projectSkills: projectSkills,
		adapters:      adapters,
	}
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

	agent := domain.Agent(req.Agent)
	now := time.Now()

	for _, sp := range skillPaths {
		conflict, err := uc.activations.FindConflict(ctx, sp.SkillName, agent, req.ProjectID)
		if err != nil {
			return fmt.Errorf("associate project category: check conflict for %s: %w", sp.SkillName, err)
		}
		if conflict != nil {
			continue // already active via global or project — skip
		}
		activation := domain.Activation{
			SkillID:   sp.SkillName,
			Agent:     agent,
			Scope:     domain.ScopeProject,
			ProjectID: &req.ProjectID,
			AppliedAt: now,
		}
		if _, err := uc.activations.Save(ctx, activation); err != nil {
			return fmt.Errorf("associate project category: save activation for %s: %w", sp.SkillName, err)
		}
	}

	adapter, ok := uc.adapters[agent]
	if !ok {
		return nil
	}
	if err := applyProjectAdapter(ctx, project, agent, uc.activations, uc.skills, uc.projectSkills, adapter); err != nil {
		return fmt.Errorf("associate project category: apply adapter: %w", err)
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
