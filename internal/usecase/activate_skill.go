package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"skill-manager/internal/domain"
)

// ActivateResult carries either the new activation or a detected conflict.
type ActivateResult struct {
	Activation *domain.Activation
	Conflict   *domain.Conflict
}

// ActivateSkillRequest holds the parameters for a skill activation.
type ActivateSkillRequest struct {
	SkillID   string
	Agent     domain.Agent
	Scope     domain.Scope
	ProjectID string // required when Scope == ScopeProject
}

// ActivateSkill activates a skill for a given agent and scope.
// When a conflict is detected it returns an ActivateResult with Conflict set
// and no error — the caller must resolve it via ResolveConflict.
type ActivateSkill struct {
	skills      SkillRepository
	projects    ProjectRepository
	activations ActivationRepository
	adapters    map[domain.Agent]AgentAdapter
}

func NewActivateSkill(
	skills SkillRepository,
	projects ProjectRepository,
	activations ActivationRepository,
	adapters map[domain.Agent]AgentAdapter,
) *ActivateSkill {
	return &ActivateSkill{
		skills:      skills,
		projects:    projects,
		activations: activations,
		adapters:    adapters,
	}
}

func (uc *ActivateSkill) Execute(ctx context.Context, req ActivateSkillRequest) (ActivateResult, error) {
	if err := uc.validateRequest(req); err != nil {
		return ActivateResult{}, err
	}

	if _, err := uc.skills.GetByID(ctx, req.SkillID); err != nil {
		if errors.Is(err, domain.ErrSkillNotFound) {
			return ActivateResult{}, domain.ErrSkillNotFound
		}
		return ActivateResult{}, fmt.Errorf("activate skill: fetch skill: %w", err)
	}

	if req.Scope == domain.ScopeProject {
		project, err := uc.projects.GetByID(ctx, req.ProjectID)
		if err != nil {
			if errors.Is(err, domain.ErrProjectNotFound) {
				return ActivateResult{}, domain.ErrProjectNotFound
			}
			return ActivateResult{}, fmt.Errorf("activate skill: fetch project: %w", err)
		}
		if !project.SupportsAgent(req.Agent) {
			return ActivateResult{}, domain.ErrAgentNotInProject
		}

		conflict, err := uc.activations.FindConflict(ctx, req.SkillID, req.Agent, req.ProjectID)
		if err != nil {
			return ActivateResult{}, fmt.Errorf("activate skill: check conflict: %w", err)
		}
		if conflict != nil {
			return ActivateResult{Conflict: conflict}, nil
		}
	}

	activation := domain.Activation{
		SkillID:   req.SkillID,
		Agent:     req.Agent,
		Scope:     req.Scope,
		AppliedAt: time.Now(),
	}
	if req.Scope == domain.ScopeProject {
		activation.ProjectID = &req.ProjectID
	}

	saved, err := uc.activations.Save(ctx, activation)
	if err != nil {
		return ActivateResult{}, fmt.Errorf("activate skill: save activation: %w", err)
	}

	if err := uc.applyAdapter(ctx, req); err != nil {
		return ActivateResult{}, fmt.Errorf("activate skill: apply adapter: %w", err)
	}

	return ActivateResult{Activation: &saved}, nil
}

func (uc *ActivateSkill) validateRequest(req ActivateSkillRequest) error {
	if req.SkillID == "" {
		return fmt.Errorf("%w: skill id is required", domain.ErrInvalidScope)
	}
	if req.Scope == domain.ScopeProject && req.ProjectID == "" {
		return fmt.Errorf("%w: project id is required for project scope", domain.ErrInvalidScope)
	}
	if req.Scope == domain.ScopeGlobal && req.ProjectID != "" {
		return fmt.Errorf("%w: project id must be empty for global scope", domain.ErrInvalidScope)
	}
	return nil
}

func (uc *ActivateSkill) applyAdapter(ctx context.Context, req ActivateSkillRequest) error {
	adapter, ok := uc.adapters[req.Agent]
	if !ok {
		return nil
	}

	allActivations, err := uc.activations.List(ctx, ActivationFilter{
		Agent: req.Agent,
		Scope: req.Scope,
	})
	if err != nil {
		return fmt.Errorf("list activations for agent: %w", err)
	}

	skillIDs := make([]string, len(allActivations))
	for i, a := range allActivations {
		skillIDs[i] = a.SkillID
	}

	activeSkills, err := uc.resolveSkills(ctx, skillIDs)
	if err != nil {
		return err
	}

	if req.Scope == domain.ScopeGlobal {
		return adapter.ApplyGlobal(ctx, activeSkills)
	}

	project, err := uc.projects.GetByID(ctx, req.ProjectID)
	if err != nil {
		return fmt.Errorf("fetch project for adapter: %w", err)
	}
	return adapter.ApplyProject(ctx, project, activeSkills)
}

func (uc *ActivateSkill) resolveSkills(ctx context.Context, ids []string) ([]domain.Skill, error) {
	skills := make([]domain.Skill, 0, len(ids))
	for _, id := range ids {
		s, err := uc.skills.GetByID(ctx, id)
		if err != nil {
			if errors.Is(err, domain.ErrSkillNotFound) {
				continue
			}
			return nil, fmt.Errorf("resolve skill %s: %w", id, err)
		}
		skills = append(skills, s)
	}
	return skills, nil
}
