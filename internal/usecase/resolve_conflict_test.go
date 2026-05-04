package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"skill-manager/internal/domain"
)

func TestResolveConflict_Execute(t *testing.T) {
	t.Parallel()

	projectID := "proj-1"
	project := domain.Project{
		ID:             projectID,
		Name:           "my-service",
		DetectedAgents: []domain.Agent{domain.AgentClaude},
	}
	skill := domain.Skill{ID: "skill-1", Name: "go-review", UpdatedAt: time.Now()}
	globalAct := &domain.Activation{
		ID: 1, SkillID: skill.ID, Agent: domain.AgentClaude, Scope: domain.ScopeGlobal,
	}
	conflict := domain.Conflict{
		SkillID:          skill.ID,
		Agent:            domain.AgentClaude,
		GlobalActivation: globalAct,
		ProjectActivation: &domain.Activation{
			SkillID: skill.ID, Agent: domain.AgentClaude, Scope: domain.ScopeProject, ProjectID: &projectID,
		},
	}

	t.Run("keep global is a no-op", func(t *testing.T) {
		t.Parallel()

		uc := NewResolveConflict(
			NewMockProjectRepository(t),
			NewMockSkillRepository(t),
			NewMockProjectSkillRepository(t),
			NewMockActivationRepository(t),
			map[domain.Agent]AgentAdapter{},
		)

		err := uc.Execute(context.Background(), ResolveConflictRequest{
			Conflict:   conflict,
			Resolution: domain.ResolutionKeepGlobal,
		})
		require.NoError(t, err)
	})

	t.Run("override project saves activation and applies adapter", func(t *testing.T) {
		t.Parallel()

		saved := domain.Activation{
			ID: 10, SkillID: skill.ID, Agent: domain.AgentClaude, Scope: domain.ScopeProject, ProjectID: &projectID,
		}
		activationRepo := NewMockActivationRepository(t)
		activationRepo.EXPECT().
			Save(context.Background(), mock.MatchedBy(func(a domain.Activation) bool {
				return a.SkillID == skill.ID && a.Scope == domain.ScopeProject && a.ProjectID != nil && *a.ProjectID == projectID
			})).
			Return(saved, nil)
		activationRepo.EXPECT().
			List(context.Background(), ActivationFilter{Agent: domain.AgentClaude, Scope: domain.ScopeProject, ProjectID: projectID}).
			Return([]domain.Activation{saved}, nil)

		projectRepo := NewMockProjectRepository(t)
		projectRepo.EXPECT().GetByID(context.Background(), projectID).Return(project, nil)

		skillRepo := NewMockSkillRepository(t)
		skillRepo.EXPECT().GetByID(context.Background(), skill.ID).Return(skill, nil)

		adapter := NewMockAgentAdapter(t)
		adapter.EXPECT().ApplyProject(context.Background(), project, []domain.Skill{skill}).Return(nil)

		uc := NewResolveConflict(
			projectRepo,
			skillRepo,
			NewMockProjectSkillRepository(t),
			activationRepo,
			map[domain.Agent]AgentAdapter{domain.AgentClaude: adapter},
		)

		err := uc.Execute(context.Background(), ResolveConflictRequest{
			Conflict:   conflict,
			Resolution: domain.ResolutionOverrideProject,
		})
		require.NoError(t, err)
	})

	t.Run("unknown resolution returns error", func(t *testing.T) {
		t.Parallel()

		uc := NewResolveConflict(
			NewMockProjectRepository(t),
			NewMockSkillRepository(t),
			NewMockProjectSkillRepository(t),
			NewMockActivationRepository(t),
			map[domain.Agent]AgentAdapter{},
		)

		err := uc.Execute(context.Background(), ResolveConflictRequest{
			Conflict:   conflict,
			Resolution: domain.ConflictResolution(99),
		})
		require.ErrorIs(t, err, domain.ErrInvalidScope)
	})
}
