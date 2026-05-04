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

func TestDoctor_Execute(t *testing.T) { //nolint:funlen
	t.Parallel()

	t.Run("reports no issues when everything is consistent", func(t *testing.T) {
		t.Parallel()

		skill := domain.Skill{ID: "s1", Name: "go-review", UpdatedAt: time.Now()}
		activation := domain.Activation{ID: 1, SkillID: skill.ID, Scope: domain.ScopeGlobal}

		activationRepo := NewMockActivationRepository(t)
		activationRepo.EXPECT().List(context.Background(), ActivationFilter{}).Return([]domain.Activation{activation}, nil)
		activationRepo.EXPECT().List(context.Background(), ActivationFilter{Scope: domain.ScopeProject}).Return(nil, nil)

		skillRepo := NewMockSkillRepository(t)
		skillRepo.EXPECT().GetByID(context.Background(), skill.ID).Return(skill, nil)

		projectRepo := NewMockProjectRepository(t)
		projectRepo.EXPECT().List(context.Background()).Return(nil, nil)

		uc := NewDoctor(skillRepo, projectRepo, activationRepo, t.TempDir())
		report, err := uc.Execute(context.Background())

		require.NoError(t, err)
		assert.Empty(t, report.Issues)
	})

	t.Run("returns error when skill check fails with non-ErrSkillNotFound", func(t *testing.T) {
		t.Parallel()

		dbErr := errors.New("timeout")
		activation := domain.Activation{ID: 5, SkillID: "s1", Scope: domain.ScopeGlobal}

		activationRepo := NewMockActivationRepository(t)
		activationRepo.EXPECT().List(context.Background(), ActivationFilter{}).Return([]domain.Activation{activation}, nil)

		skillRepo := NewMockSkillRepository(t)
		skillRepo.EXPECT().GetByID(context.Background(), "s1").Return(domain.Skill{}, dbErr)

		uc := NewDoctor(skillRepo, NewMockProjectRepository(t), activationRepo, t.TempDir())
		_, err := uc.Execute(context.Background())

		require.ErrorIs(t, err, dbErr)
	})

	t.Run("reports orphaned activation when skill is missing", func(t *testing.T) {
		t.Parallel()

		activation := domain.Activation{ID: 7, SkillID: "gone-skill", Scope: domain.ScopeGlobal}

		activationRepo := NewMockActivationRepository(t)
		activationRepo.EXPECT().List(context.Background(), ActivationFilter{}).Return([]domain.Activation{activation}, nil)
		activationRepo.EXPECT().List(context.Background(), ActivationFilter{Scope: domain.ScopeProject}).Return(nil, nil)

		skillRepo := NewMockSkillRepository(t)
		skillRepo.EXPECT().GetByID(context.Background(), "gone-skill").Return(domain.Skill{}, domain.ErrSkillNotFound)

		projectRepo := NewMockProjectRepository(t)
		projectRepo.EXPECT().List(context.Background()).Return(nil, nil)

		uc := NewDoctor(skillRepo, projectRepo, activationRepo, t.TempDir())
		report, err := uc.Execute(context.Background())

		require.NoError(t, err)
		require.Len(t, report.Issues, 1)
		assert.Equal(t, "orphaned_activation", report.Issues[0].Kind)
		assert.True(t, report.Issues[0].Fixable)
	})

	t.Run("returns error when list activations fails", func(t *testing.T) {
		t.Parallel()

		dbErr := errors.New("db connection lost")
		activationRepo := NewMockActivationRepository(t)
		activationRepo.EXPECT().List(context.Background(), ActivationFilter{}).Return(nil, dbErr)

		uc := NewDoctor(NewMockSkillRepository(t), NewMockProjectRepository(t), activationRepo, t.TempDir())
		_, err := uc.Execute(context.Background())

		require.ErrorIs(t, err, dbErr)
	})

	t.Run("returns error when list project activations fails", func(t *testing.T) {
		t.Parallel()

		dbErr := errors.New("db error")
		activationRepo := NewMockActivationRepository(t)
		activationRepo.EXPECT().List(context.Background(), ActivationFilter{}).Return(nil, nil)
		activationRepo.EXPECT().List(context.Background(), ActivationFilter{Scope: domain.ScopeProject}).Return(nil, dbErr)

		uc := NewDoctor(NewMockSkillRepository(t), NewMockProjectRepository(t), activationRepo, t.TempDir())
		_, err := uc.Execute(context.Background())

		require.ErrorIs(t, err, dbErr)
	})

	t.Run("reports missing project when project activation references deleted project", func(t *testing.T) {
		t.Parallel()

		projID := "deleted-proj"
		activation := domain.Activation{
			ID: 3, SkillID: "s1", Scope: domain.ScopeProject, ProjectID: &projID,
		}

		activationRepo := NewMockActivationRepository(t)
		activationRepo.EXPECT().List(context.Background(), ActivationFilter{}).Return([]domain.Activation{activation}, nil)
		activationRepo.EXPECT().List(context.Background(), ActivationFilter{Scope: domain.ScopeProject}).Return([]domain.Activation{activation}, nil)

		skillRepo := NewMockSkillRepository(t)
		skillRepo.EXPECT().GetByID(context.Background(), "s1").Return(domain.Skill{ID: "s1"}, nil)

		projectRepo := NewMockProjectRepository(t)
		projectRepo.EXPECT().GetByID(context.Background(), projID).Return(domain.Project{}, domain.ErrProjectNotFound)

		projectRepo.EXPECT().List(context.Background()).Return(nil, nil)

		uc := NewDoctor(skillRepo, projectRepo, activationRepo, t.TempDir())
		report, err := uc.Execute(context.Background())

		require.NoError(t, err)
		require.Len(t, report.Issues, 1)
		assert.Equal(t, "missing_project", report.Issues[0].Kind)
	})
}
