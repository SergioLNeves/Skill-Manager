package usecase

import (
	"context"
	"fmt"
	"time"

	"skill-manager/internal/domain"
)

// AssignSkillCategory sets (or clears) the category for a skill and propagates
// the change: for each project already linked to the new category (and agent),
// creates an Activation for the skill and syncs the adapter.
type AssignSkillCategory struct {
	categories    CategoryRepository
	projects      ProjectRepository
	activations   ActivationRepository
	skills        SkillRepository
	projectSkills ProjectSkillRepository
	adapters      map[domain.Agent]AgentAdapter
}

func NewAssignSkillCategory(
	categories CategoryRepository,
	projects ProjectRepository,
	activations ActivationRepository,
	skills SkillRepository,
	projectSkills ProjectSkillRepository,
	adapters map[domain.Agent]AgentAdapter,
) *AssignSkillCategory {
	return &AssignSkillCategory{
		categories:    categories,
		projects:      projects,
		activations:   activations,
		skills:        skills,
		projectSkills: projectSkills,
		adapters:      adapters,
	}
}

type AssignSkillCategoryRequest struct {
	SkillName  string
	SkillPath  string // filesystem path of the skill directory
	CategoryID *int64 // nil to unassign
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

	now := time.Now()

	for _, link := range links {
		project, err := uc.projects.GetByID(ctx, link.ProjectID)
		if err != nil {
			continue
		}

		agent := link.Agent

		conflict, err := uc.activations.FindConflict(ctx, req.SkillName, agent, link.ProjectID)
		if err != nil {
			continue
		}
		if conflict != nil {
			continue // already active — skip
		}

		projectID := link.ProjectID
		activation := domain.Activation{
			SkillID:   req.SkillName,
			Agent:     agent,
			Scope:     domain.ScopeProject,
			ProjectID: &projectID,
			AppliedAt: now,
		}
		if _, err := uc.activations.Save(ctx, activation); err != nil {
			continue // best-effort per project
		}

		adapter, ok := uc.adapters[agent]
		if !ok {
			continue
		}
		// Sync adapter once per (project, agent) — errors are best-effort
		_ = applyProjectAdapter(ctx, project, agent, uc.activations, uc.skills, uc.projectSkills, adapter)
	}

	return nil
}
