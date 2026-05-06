package usecase

import (
	"context"
	"errors"
	"fmt"

	"skill-manager/internal/domain"
)

// applyProjectAdapter collects all active skills for a project+agent, resolves their
// paths, and calls adapter.ApplyProject — keeping the adapter's filesystem state in
// sync with the activations table.
func applyProjectAdapter(
	ctx context.Context,
	project domain.Project,
	agent domain.Agent,
	activations ActivationRepository,
	skills SkillRepository,
	projectSkills ProjectSkillRepository,
	adapter AgentAdapter,
) error {
	allActivations, err := activations.List(ctx, ActivationFilter{
		Agent:     agent,
		Scope:     domain.ScopeProject,
		ProjectID: project.ID,
	})
	if err != nil {
		return fmt.Errorf("applyProjectAdapter: list activations: %w", err)
	}

	activeSkills := make([]domain.Skill, 0, len(allActivations))
	for _, a := range allActivations {
		s, err := resolveSkillForProject(ctx, a.SkillID, project, skills, projectSkills)
		if err != nil {
			if errors.Is(err, domain.ErrSkillNotFound) {
				continue
			}
			return fmt.Errorf("applyProjectAdapter: resolve skill %s: %w", a.SkillID, err)
		}
		activeSkills = append(activeSkills, s)
	}

	return adapter.ApplyProject(ctx, project, activeSkills)
}

// resolveSkillForProject looks up a skill by ID/name in the global repo first,
// then falls back to the project skill repo.
func resolveSkillForProject(
	ctx context.Context,
	skillID string,
	project domain.Project,
	skills SkillRepository,
	projectSkills ProjectSkillRepository,
) (domain.Skill, error) {
	s, err := skills.GetByID(ctx, skillID)
	if err == nil {
		return s, nil
	}
	if !errors.Is(err, domain.ErrSkillNotFound) {
		return domain.Skill{}, err
	}
	s, err = projectSkills.GetByID(ctx, skillID, project)
	if err != nil {
		if errors.Is(err, domain.ErrSkillNotFound) {
			return domain.Skill{}, domain.ErrSkillNotFound
		}
		return domain.Skill{}, err
	}
	return s, nil
}
