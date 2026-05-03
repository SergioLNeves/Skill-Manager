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

func TestListSkills_Execute(t *testing.T) {
	t.Parallel()

	t.Run("returns skills from repository", func(t *testing.T) {
		t.Parallel()

		skills := []domain.Skill{
			{ID: "1", Name: "go-review", Description: "Go code review", UpdatedAt: time.Now()},
			{ID: "2", Name: "pt-br", Description: "Portuguese style", UpdatedAt: time.Now()},
		}
		repo := NewMockSkillRepository(t)
		repo.EXPECT().List(context.Background()).Return(skills, nil)

		uc := NewListSkills(repo)
		got, err := uc.Execute(context.Background())

		require.NoError(t, err)
		assert.Equal(t, skills, got)
	})

	t.Run("wraps repository error", func(t *testing.T) {
		t.Parallel()

		repoErr := errors.New("disk error")
		repo := NewMockSkillRepository(t)
		repo.EXPECT().List(context.Background()).Return(nil, repoErr)

		uc := NewListSkills(repo)
		_, err := uc.Execute(context.Background())

		require.Error(t, err)
		assert.ErrorIs(t, err, repoErr)
	})
}
