package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"skill-manager/internal/domain"
)

func TestActivateSkill_Execute(t *testing.T) {
	t.Parallel()

	skill := domain.Skill{ID: "skill-1", Name: "go-review", UpdatedAt: time.Now()}
	project := domain.Project{
		ID:             "proj-1",
		Name:           "my-service",
		Path:           "/home/user/my-service",
		DetectedAgents: []domain.Agent{domain.AgentClaude},
	}
	projectID := project.ID

	t.Run("global activation happy path", func(t *testing.T) {
		t.Parallel()

		saved := domain.Activation{
			ID: 1, SkillID: skill.ID, Agent: domain.AgentClaude, Scope: domain.ScopeGlobal, AppliedAt: time.Now(),
		}

		skillRepo := NewMockSkillRepository(t)
		skillRepo.EXPECT().GetByID(context.Background(), skill.ID).Return(skill, nil)

		activationRepo := NewMockActivationRepository(t)
		activationRepo.EXPECT().
			Save(context.Background(), mock.MatchedBy(func(a domain.Activation) bool {
				return a.SkillID == skill.ID && a.Agent == domain.AgentClaude && a.Scope == domain.ScopeGlobal
			})).
			RunAndReturn(func(_ context.Context, a domain.Activation) (domain.Activation, error) {
				a.ID = 1
				return a, nil
			})
		activationRepo.EXPECT().
			List(context.Background(), ActivationFilter{Agent: domain.AgentClaude, Scope: domain.ScopeGlobal}).
			Return([]domain.Activation{saved}, nil)

		skillRepo.EXPECT().GetByID(context.Background(), skill.ID).Return(skill, nil)

		adapter := NewMockAgentAdapter(t)
		adapter.EXPECT().Agent().Return(domain.AgentClaude).Maybe()
		adapter.EXPECT().ApplyGlobal(context.Background(), []domain.Skill{skill}).Return(nil)

		uc := NewActivateSkill(skillRepo, NewMockProjectSkillRepository(t), NewMockProjectRepository(t), activationRepo,
			map[domain.Agent]AgentAdapter{domain.AgentClaude: adapter})

		result, err := uc.Execute(context.Background(), ActivateSkillRequest{
			SkillID: skill.ID,
			Agent:   domain.AgentClaude,
			Scope:   domain.ScopeGlobal,
		})

		require.NoError(t, err)
		require.NotNil(t, result.Activation)
		assert.Nil(t, result.Conflict)
	})

	t.Run("project activation returns conflict when global exists", func(t *testing.T) {
		t.Parallel()

		globalActivation := &domain.Activation{
			ID: 1, SkillID: skill.ID, Agent: domain.AgentClaude, Scope: domain.ScopeGlobal,
		}
		conflict := &domain.Conflict{
			SkillID:          skill.ID,
			Agent:            domain.AgentClaude,
			GlobalActivation: globalActivation,
		}

		skillRepo := NewMockSkillRepository(t)
		skillRepo.EXPECT().GetByID(context.Background(), skill.ID).Return(skill, nil)

		projectRepo := NewMockProjectRepository(t)
		projectRepo.EXPECT().GetByID(context.Background(), projectID).Return(project, nil)

		activationRepo := NewMockActivationRepository(t)
		activationRepo.EXPECT().
			FindConflict(context.Background(), skill.ID, domain.AgentClaude, projectID).
			Return(conflict, nil)

		uc := NewActivateSkill(skillRepo, NewMockProjectSkillRepository(t), projectRepo, activationRepo,
			map[domain.Agent]AgentAdapter{})

		result, err := uc.Execute(context.Background(), ActivateSkillRequest{
			SkillID:   skill.ID,
			Agent:     domain.AgentClaude,
			Scope:     domain.ScopeProject,
			ProjectID: projectID,
		})

		require.NoError(t, err)
		require.NotNil(t, result.Conflict)
		assert.Nil(t, result.Activation)
	})

	t.Run("returns ErrSkillNotFound when skill missing", func(t *testing.T) {
		t.Parallel()

		skillRepo := NewMockSkillRepository(t)
		skillRepo.EXPECT().GetByID(context.Background(), "missing").Return(domain.Skill{}, domain.ErrSkillNotFound)

		uc := NewActivateSkill(skillRepo, NewMockProjectSkillRepository(t), NewMockProjectRepository(t), NewMockActivationRepository(t),
			map[domain.Agent]AgentAdapter{})

		_, err := uc.Execute(context.Background(), ActivateSkillRequest{
			SkillID: "missing",
			Agent:   domain.AgentClaude,
			Scope:   domain.ScopeGlobal,
		})

		require.ErrorIs(t, err, domain.ErrSkillNotFound)
	})

	t.Run("returns ErrInvalidScope when project id missing for project scope", func(t *testing.T) {
		t.Parallel()

		uc := NewActivateSkill(NewMockSkillRepository(t), NewMockProjectSkillRepository(t), NewMockProjectRepository(t),
			NewMockActivationRepository(t), map[domain.Agent]AgentAdapter{})

		_, err := uc.Execute(context.Background(), ActivateSkillRequest{
			SkillID: skill.ID,
			Agent:   domain.AgentClaude,
			Scope:   domain.ScopeProject,
		})

		require.ErrorIs(t, err, domain.ErrInvalidScope)
	})

	t.Run("returns ErrInvalidScope when project id set on global scope", func(t *testing.T) {
		t.Parallel()

		uc := NewActivateSkill(NewMockSkillRepository(t), NewMockProjectSkillRepository(t), NewMockProjectRepository(t),
			NewMockActivationRepository(t), map[domain.Agent]AgentAdapter{})

		_, err := uc.Execute(context.Background(), ActivateSkillRequest{
			SkillID:   skill.ID,
			Agent:     domain.AgentClaude,
			Scope:     domain.ScopeGlobal,
			ProjectID: "should-not-be-here",
		})

		require.ErrorIs(t, err, domain.ErrInvalidScope)
	})

	t.Run("returns ErrInvalidScope when skill id is empty", func(t *testing.T) {
		t.Parallel()

		uc := NewActivateSkill(NewMockSkillRepository(t), NewMockProjectSkillRepository(t), NewMockProjectRepository(t),
			NewMockActivationRepository(t), map[domain.Agent]AgentAdapter{})

		_, err := uc.Execute(context.Background(), ActivateSkillRequest{
			Agent: domain.AgentClaude,
			Scope: domain.ScopeGlobal,
		})

		require.ErrorIs(t, err, domain.ErrInvalidScope)
	})

	t.Run("returns ErrProjectNotFound when project missing", func(t *testing.T) {
		t.Parallel()

		skillRepo := NewMockSkillRepository(t)
		skillRepo.EXPECT().GetByID(context.Background(), skill.ID).Return(skill, nil)

		projectRepo := NewMockProjectRepository(t)
		projectRepo.EXPECT().GetByID(context.Background(), "missing-proj").Return(domain.Project{}, domain.ErrProjectNotFound)

		uc := NewActivateSkill(skillRepo, NewMockProjectSkillRepository(t), projectRepo, NewMockActivationRepository(t),
			map[domain.Agent]AgentAdapter{})

		_, err := uc.Execute(context.Background(), ActivateSkillRequest{
			SkillID:   skill.ID,
			Agent:     domain.AgentClaude,
			Scope:     domain.ScopeProject,
			ProjectID: "missing-proj",
		})

		require.ErrorIs(t, err, domain.ErrProjectNotFound)
	})

	t.Run("skips missing skill in resolveSkills during applyAdapter", func(t *testing.T) {
		t.Parallel()

		saved := domain.Activation{
			ID: 3, SkillID: skill.ID, Agent: domain.AgentClaude, Scope: domain.ScopeGlobal,
		}
		missingActivation := domain.Activation{
			ID: 4, SkillID: "gone", Agent: domain.AgentClaude, Scope: domain.ScopeGlobal,
		}

		skillRepo := NewMockSkillRepository(t)
		skillRepo.EXPECT().GetByID(context.Background(), skill.ID).Return(skill, nil)
		skillRepo.EXPECT().GetByID(context.Background(), skill.ID).Return(skill, nil)
		skillRepo.EXPECT().GetByID(context.Background(), "gone").Return(domain.Skill{}, domain.ErrSkillNotFound)

		activationRepo := NewMockActivationRepository(t)
		activationRepo.EXPECT().
			Save(context.Background(), mock.MatchedBy(func(a domain.Activation) bool {
				return a.SkillID == skill.ID && a.Scope == domain.ScopeGlobal
			})).
			Return(saved, nil)
		activationRepo.EXPECT().
			List(context.Background(), ActivationFilter{Agent: domain.AgentClaude, Scope: domain.ScopeGlobal}).
			Return([]domain.Activation{saved, missingActivation}, nil)

		adapter := NewMockAgentAdapter(t)
		adapter.EXPECT().ApplyGlobal(context.Background(), []domain.Skill{skill}).Return(nil)

		uc := NewActivateSkill(skillRepo, NewMockProjectSkillRepository(t), NewMockProjectRepository(t), activationRepo,
			map[domain.Agent]AgentAdapter{domain.AgentClaude: adapter})

		result, err := uc.Execute(context.Background(), ActivateSkillRequest{
			SkillID: skill.ID,
			Agent:   domain.AgentClaude,
			Scope:   domain.ScopeGlobal,
		})

		require.NoError(t, err)
		require.NotNil(t, result.Activation)
	})

	t.Run("project activation happy path calls ApplyProject", func(t *testing.T) {
		t.Parallel()

		pid := projectID
		saved := domain.Activation{
			ID: 2, SkillID: skill.ID, Agent: domain.AgentClaude, Scope: domain.ScopeProject, ProjectID: &pid,
		}

		skillRepo := NewMockSkillRepository(t)
		skillRepo.EXPECT().GetByID(context.Background(), skill.ID).Return(skill, nil)
		skillRepo.EXPECT().GetByID(context.Background(), skill.ID).Return(skill, nil)

		projectRepo := NewMockProjectRepository(t)
		projectRepo.EXPECT().GetByID(context.Background(), pid).Return(project, nil)
		projectRepo.EXPECT().GetByID(context.Background(), pid).Return(project, nil)

		activationRepo := NewMockActivationRepository(t)
		activationRepo.EXPECT().
			FindConflict(context.Background(), skill.ID, domain.AgentClaude, pid).
			Return(nil, nil)
		activationRepo.EXPECT().
			Save(context.Background(), mock.MatchedBy(func(a domain.Activation) bool {
				return a.SkillID == skill.ID && a.Scope == domain.ScopeProject
			})).
			Return(saved, nil)
		activationRepo.EXPECT().
			List(context.Background(), ActivationFilter{Agent: domain.AgentClaude, Scope: domain.ScopeProject}).
			Return([]domain.Activation{saved}, nil)

		adapter := NewMockAgentAdapter(t)
		adapter.EXPECT().ApplyProject(context.Background(), project, []domain.Skill{skill}).Return(nil)

		uc := NewActivateSkill(skillRepo, NewMockProjectSkillRepository(t), projectRepo, activationRepo,
			map[domain.Agent]AgentAdapter{domain.AgentClaude: adapter})

		result, err := uc.Execute(context.Background(), ActivateSkillRequest{
			SkillID:   skill.ID,
			Agent:     domain.AgentClaude,
			Scope:     domain.ScopeProject,
			ProjectID: pid,
		})

		require.NoError(t, err)
		require.NotNil(t, result.Activation)
		assert.Nil(t, result.Conflict)
	})
}
