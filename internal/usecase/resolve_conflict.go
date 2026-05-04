package usecase

import (
	"context"
	"fmt"
	"time"

	"skill-manager/internal/domain"
)

// ResolveConflictRequest carries the conflict and the chosen resolution.
type ResolveConflictRequest struct {
	Conflict   domain.Conflict
	Resolution domain.ConflictResolution
}

// ResolveConflict applies the user's chosen resolution to an activation conflict.
type ResolveConflict struct {
	projects      ProjectRepository
	skills        SkillRepository
	projectSkills ProjectSkillRepository
	activations   ActivationRepository
	adapters      map[domain.Agent]AgentAdapter
}

func NewResolveConflict(
	projects ProjectRepository,
	skills SkillRepository,
	projectSkills ProjectSkillRepository,
	activations ActivationRepository,
	adapters map[domain.Agent]AgentAdapter,
) *ResolveConflict {
	return &ResolveConflict{
		projects:      projects,
		skills:        skills,
		projectSkills: projectSkills,
		activations:   activations,
		adapters:      adapters,
	}
}

func (uc *ResolveConflict) Execute(ctx context.Context, req ResolveConflictRequest) error {
	switch req.Resolution {
	case domain.ResolutionKeepGlobal:
		return nil // no-op: global stays, project activation is simply discarded

	case domain.ResolutionOverrideProject:
		return uc.saveProjectOverride(ctx, req.Conflict)

	default:
		return fmt.Errorf("%w: unknown resolution %d", domain.ErrInvalidScope, req.Resolution)
	}
}

func (uc *ResolveConflict) saveProjectOverride(ctx context.Context, c domain.Conflict) error {
	if c.GlobalActivation == nil || c.GlobalActivation.ProjectID == nil {
		// this is intentional: global has no project id, we want project's
	}

	projectID := ""
	if c.ProjectActivation != nil && c.ProjectActivation.ProjectID != nil {
		projectID = *c.ProjectActivation.ProjectID
	}

	activation := domain.Activation{
		SkillID:   c.SkillID,
		Agent:     c.Agent,
		Scope:     domain.ScopeProject,
		ProjectID: &projectID,
		AppliedAt: time.Now(),
	}

	if _, err := uc.activations.Save(ctx, activation); err != nil {
		return fmt.Errorf("resolve conflict: save override: %w", err)
	}

	adapter, ok := uc.adapters[c.Agent]
	if !ok {
		return nil
	}

	project, err := uc.projects.GetByID(ctx, projectID)
	if err != nil {
		return fmt.Errorf("resolve conflict: fetch project: %w", err)
	}

	all, err := uc.activations.List(ctx, ActivationFilter{
		Agent:     c.Agent,
		Scope:     domain.ScopeProject,
		ProjectID: projectID,
	})
	if err != nil {
		return fmt.Errorf("resolve conflict: list activations: %w", err)
	}

	activeSkills := make([]domain.Skill, 0, len(all))
	for _, a := range all {
		s, err := uc.skills.GetByID(ctx, a.SkillID)
		if err != nil {
			// Try project skills as fallback.
			s, err = uc.projectSkills.GetByID(ctx, a.SkillID, project)
			if err != nil {
				continue
			}
		}
		activeSkills = append(activeSkills, s)
	}

	return adapter.ApplyProject(ctx, project, activeSkills)
}
