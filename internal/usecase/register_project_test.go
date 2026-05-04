package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"skill-manager/internal/domain"
)

func TestRegisterProject_Execute(t *testing.T) {
	t.Parallel()

	project := domain.Project{
		ID:      "proj-1",
		Name:    "my-service",
		Path:    "/home/user/my-service",
		AddedAt: time.Now(),
	}

	t.Run("saves project successfully", func(t *testing.T) {
		t.Parallel()

		repo := NewMockProjectRepository(t)
		repo.EXPECT().Save(context.Background(), project).Return(nil)

		uc := NewRegisterProject(repo)
		err := uc.Execute(context.Background(), project)

		require.NoError(t, err)
	})

	t.Run("wraps repository error", func(t *testing.T) {
		t.Parallel()

		repoErr := errors.New("constraint violation")
		repo := NewMockProjectRepository(t)
		repo.EXPECT().Save(context.Background(), project).Return(repoErr)

		uc := NewRegisterProject(repo)
		err := uc.Execute(context.Background(), project)

		require.ErrorIs(t, err, repoErr)
	})
}
