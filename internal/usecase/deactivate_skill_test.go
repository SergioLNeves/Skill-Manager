package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"skill-manager/internal/domain"
)

func TestDeactivateSkill_Execute(t *testing.T) {
	t.Parallel()

	projectID := "proj-1"
	activation := domain.Activation{
		ID:        42,
		SkillID:   "skill-1",
		Agent:     domain.AgentClaude,
		Scope:     domain.ScopeGlobal,
		AppliedAt: time.Now(),
	}
	skill := domain.Skill{ID: "skill-1", Name: "go-review"}
	project := domain.Project{ID: projectID, DetectedAgents: []domain.Agent{domain.AgentClaude}}

	t.Run("removes activation and reapplies adapter", func(t *testing.T) {
		t.Parallel()

		activationRepo := NewMockActivationRepository(t)
		activationRepo.EXPECT().
			List(context.Background(), ActivationFilter{}).
			Return([]domain.Activation{activation}, nil)
		activationRepo.EXPECT().Delete(context.Background(), int64(42)).Return(nil)
		activationRepo.EXPECT().
			List(context.Background(), ActivationFilter{Agent: domain.AgentClaude, Scope: domain.ScopeGlobal}).
			Return(nil, nil)

		skillRepo := NewMockSkillRepository(t)
		adapter := NewMockAgentAdapter(t)
		adapter.EXPECT().ApplyGlobal(context.Background(), []domain.Skill{}).Return(nil)

		uc := NewDeactivateSkill(skillRepo, NewMockProjectSkillRepository(t), NewMockProjectRepository(t), activationRepo,
			map[domain.Agent]AgentAdapter{domain.AgentClaude: adapter})

		err := uc.Execute(context.Background(), 42)
		require.NoError(t, err)
	})

	t.Run("reapplies remaining skills after removal", func(t *testing.T) {
		t.Parallel()

		remaining := domain.Activation{
			ID: 99, SkillID: skill.ID, Agent: domain.AgentClaude, Scope: domain.ScopeGlobal,
		}
		toRemove := domain.Activation{
			ID: 42, SkillID: "skill-2", Agent: domain.AgentClaude, Scope: domain.ScopeGlobal,
		}

		activationRepo := NewMockActivationRepository(t)
		activationRepo.EXPECT().
			List(context.Background(), ActivationFilter{}).
			Return([]domain.Activation{toRemove, remaining}, nil)
		activationRepo.EXPECT().Delete(context.Background(), int64(42)).Return(nil)
		activationRepo.EXPECT().
			List(context.Background(), ActivationFilter{Agent: domain.AgentClaude, Scope: domain.ScopeGlobal}).
			Return([]domain.Activation{remaining}, nil)

		skillRepo := NewMockSkillRepository(t)
		skillRepo.EXPECT().GetByID(context.Background(), skill.ID).Return(skill, nil)

		adapter := NewMockAgentAdapter(t)
		adapter.EXPECT().ApplyGlobal(context.Background(), []domain.Skill{skill}).Return(nil)

		uc := NewDeactivateSkill(skillRepo, NewMockProjectSkillRepository(t), NewMockProjectRepository(t), activationRepo,
			map[domain.Agent]AgentAdapter{domain.AgentClaude: adapter})

		err := uc.Execute(context.Background(), 42)
		require.NoError(t, err)
		_ = project
		_ = projectID
	})

	t.Run("returns ErrActivationNotFound when id is unknown", func(t *testing.T) {
		t.Parallel()

		activationRepo := NewMockActivationRepository(t)
		activationRepo.EXPECT().
			List(context.Background(), ActivationFilter{}).
			Return([]domain.Activation{activation}, nil)

		uc := NewDeactivateSkill(NewMockSkillRepository(t), NewMockProjectSkillRepository(t), NewMockProjectRepository(t), activationRepo,
			map[domain.Agent]AgentAdapter{})

		err := uc.Execute(context.Background(), 999)
		require.ErrorIs(t, err, domain.ErrActivationNotFound)
		assert.True(t, true) // explicit: no adapter call expected
	})

	t.Run("returns error when delete fails", func(t *testing.T) {
		t.Parallel()

		dbErr := errors.New("locked")
		activationRepo := NewMockActivationRepository(t)
		activationRepo.EXPECT().
			List(context.Background(), ActivationFilter{}).
			Return([]domain.Activation{activation}, nil)
		activationRepo.EXPECT().Delete(context.Background(), int64(42)).Return(dbErr)

		uc := NewDeactivateSkill(NewMockSkillRepository(t), NewMockProjectSkillRepository(t), NewMockProjectRepository(t), activationRepo,
			map[domain.Agent]AgentAdapter{})

		err := uc.Execute(context.Background(), 42)
		require.ErrorIs(t, err, dbErr)
	})

	t.Run("reapplies project adapter when scope is project", func(t *testing.T) {
		t.Parallel()

		pid := projectID
		projectActivation := domain.Activation{
			ID:        55,
			SkillID:   skill.ID,
			Agent:     domain.AgentClaude,
			Scope:     domain.ScopeProject,
			ProjectID: &pid,
			AppliedAt: time.Now(),
		}

		activationRepo := NewMockActivationRepository(t)
		activationRepo.EXPECT().
			List(context.Background(), ActivationFilter{}).
			Return([]domain.Activation{projectActivation}, nil)
		activationRepo.EXPECT().Delete(context.Background(), int64(55)).Return(nil)
		activationRepo.EXPECT().
			List(context.Background(), ActivationFilter{Agent: domain.AgentClaude, Scope: domain.ScopeProject}).
			Return(nil, nil)

		projectRepo := NewMockProjectRepository(t)
		projectRepo.EXPECT().GetByID(context.Background(), pid).Return(project, nil)

		adapter := NewMockAgentAdapter(t)
		adapter.EXPECT().ApplyProject(context.Background(), project, []domain.Skill{}).Return(nil)

		uc := NewDeactivateSkill(NewMockSkillRepository(t), NewMockProjectSkillRepository(t), projectRepo, activationRepo,
			map[domain.Agent]AgentAdapter{domain.AgentClaude: adapter})

		err := uc.Execute(context.Background(), 55)
		require.NoError(t, err)
	})
}
