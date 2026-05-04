package usecase

import (
	"context"
	"errors"
	"fmt"

	"skill-manager/internal/domain"
)

// DeactivateSkill removes an activation and re-applies the adapter state.
type DeactivateSkill struct {
	skills      SkillRepository
	projects    ProjectRepository
	activations ActivationRepository
	adapters    map[domain.Agent]AgentAdapter
}

func NewDeactivateSkill(
	skills SkillRepository,
	projects ProjectRepository,
	activations ActivationRepository,
	adapters map[domain.Agent]AgentAdapter,
) *DeactivateSkill {
	return &DeactivateSkill{
		skills:      skills,
		projects:    projects,
		activations: activations,
		adapters:    adapters,
	}
}

func (uc *DeactivateSkill) Execute(ctx context.Context, activationID int64) error {
	// Fetch first so we know agent/scope/projectID for reapply before deleting.
	all, err := uc.activations.List(ctx, ActivationFilter{})
	if err != nil {
		return fmt.Errorf("deactivate skill: list activations: %w", err)
	}
	var target *domain.Activation
	for i := range all {
		if all[i].ID == activationID {
			target = &all[i]
			break
		}
	}
	if target == nil {
		return domain.ErrActivationNotFound
	}

	if err := uc.activations.Delete(ctx, activationID); err != nil {
		return fmt.Errorf("deactivate skill: delete activation: %w", err)
	}
	if err := uc.reapply(ctx, target); err != nil {
		return fmt.Errorf("deactivate skill: reapply adapter: %w", err)
	}
	return nil
}

func (uc *DeactivateSkill) reapply(ctx context.Context, removed *domain.Activation) error {
	adapter, ok := uc.adapters[removed.Agent]
	if !ok {
		return nil
	}

	remaining, err := uc.activations.List(ctx, ActivationFilter{
		Agent: removed.Agent,
		Scope: removed.Scope,
	})
	if err != nil {
		return fmt.Errorf("list remaining activations: %w", err)
	}

	skillIDs := make([]string, len(remaining))
	for i, a := range remaining {
		skillIDs[i] = a.SkillID
	}

	activeSkills := make([]domain.Skill, 0, len(skillIDs))
	for _, id := range skillIDs {
		s, err := uc.skills.GetByID(ctx, id)
		if err != nil {
			if errors.Is(err, domain.ErrSkillNotFound) {
				continue
			}
			return fmt.Errorf("resolve skill %s: %w", id, err)
		}
		activeSkills = append(activeSkills, s)
	}

	if removed.Scope == domain.ScopeGlobal {
		return adapter.ApplyGlobal(ctx, activeSkills)
	}

	project, err := uc.projects.GetByID(ctx, *removed.ProjectID)
	if err != nil {
		return fmt.Errorf("fetch project for adapter: %w", err)
	}
	return adapter.ApplyProject(ctx, project, activeSkills)
}
