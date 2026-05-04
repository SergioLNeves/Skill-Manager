package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"skill-manager/internal/domain"
)

func TestScanProjects_Execute(t *testing.T) {
	t.Parallel()

	candidates := []ProjectCandidate{
		{Name: "service-a", Path: "/home/user/service-a", DetectedAgents: []domain.Agent{domain.AgentClaude}},
		{Name: "service-b", Path: "/home/user/service-b", DetectedAgents: []domain.Agent{domain.AgentCopilot}},
	}
	roots := []string{"/home/user"}

	t.Run("returns candidates from scanner", func(t *testing.T) {
		t.Parallel()

		scanner := NewMockProjectScanner(t)
		scanner.EXPECT().Scan(context.Background(), roots).Return(candidates, nil)

		uc := NewScanProjects(scanner)
		got, err := uc.Execute(context.Background(), roots)

		require.NoError(t, err)
		assert.Equal(t, candidates, got)
	})

	t.Run("wraps scanner error", func(t *testing.T) {
		t.Parallel()

		scanErr := errors.New("permission denied")
		scanner := NewMockProjectScanner(t)
		scanner.EXPECT().Scan(context.Background(), roots).Return(nil, scanErr)

		uc := NewScanProjects(scanner)
		_, err := uc.Execute(context.Background(), roots)

		require.ErrorIs(t, err, scanErr)
	})
}
