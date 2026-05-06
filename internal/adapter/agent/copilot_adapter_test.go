package agent

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"skill-manager/internal/domain"
)

func TestCopilotAdapter_Agent(t *testing.T) {
	a := NewCopilotAdapter()
	assert.Equal(t, domain.AgentCopilot, a.Agent())
}

func TestCopilotAdapter_CapabilityCheck(t *testing.T) {
	a := NewCopilotAdapter()
	assert.NoError(t, a.CapabilityCheck(context.Background()))
}

func TestCopilotAdapter_ApplyGlobal(t *testing.T) {
	a := NewCopilotAdapter()
	// Global is always a no-op.
	err := a.ApplyGlobal(context.Background(), []domain.Skill{{Name: "x"}})
	assert.NoError(t, err)
}

func makeTestSkills(t *testing.T) []domain.Skill {
	t.Helper()
	base := t.TempDir()
	skills := []domain.Skill{
		{Name: "go-review", Description: "Go code review checklist"},
		{Name: "debug-tips", Description: "Debugging tips"},
	}
	for i, s := range skills {
		dir := filepath.Join(base, s.Name)
		require.NoError(t, os.MkdirAll(dir, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("# "+s.Name), 0o644))
		skills[i].Path = dir
	}
	return skills
}

func TestCopilotAdapter_ApplyProject(t *testing.T) {
	t.Parallel()

	skills := makeTestSkills(t)

	t.Run("creates file with managed block when file does not exist", func(t *testing.T) {
		t.Parallel()

		projectDir := t.TempDir()
		a := NewCopilotAdapter()

		require.NoError(t, a.ApplyProject(context.Background(), domain.Project{Path: projectDir}, skills))

		content, err := os.ReadFile(filepath.Join(projectDir, copilotFilePath))
		require.NoError(t, err)
		s := string(content)
		assert.Contains(t, s, copilotStartMarker)
		assert.Contains(t, s, copilotEndMarker)
		assert.Contains(t, s, "go-review")
		assert.Contains(t, s, "debug-tips")
	})

	t.Run("preserves user content before managed block", func(t *testing.T) {
		t.Parallel()

		projectDir := t.TempDir()
		instructionsPath := filepath.Join(projectDir, copilotFilePath)
		require.NoError(t, os.MkdirAll(filepath.Dir(instructionsPath), 0o755))
		existing := "# My Project\n\nSome user instructions.\n\n" +
			copilotStartMarker + "\nold content\n" + copilotEndMarker
		require.NoError(t, os.WriteFile(instructionsPath, []byte(existing), 0o644))

		a := NewCopilotAdapter()
		require.NoError(t, a.ApplyProject(context.Background(), domain.Project{Path: projectDir}, skills))

		content, _ := os.ReadFile(instructionsPath)
		s := string(content)
		assert.True(t, strings.HasPrefix(s, "# My Project"))
		assert.Contains(t, s, "Some user instructions.")
		assert.Contains(t, s, "go-review")
	})

	t.Run("preserves user content after managed block", func(t *testing.T) {
		t.Parallel()

		projectDir := t.TempDir()
		instructionsPath := filepath.Join(projectDir, copilotFilePath)
		require.NoError(t, os.MkdirAll(filepath.Dir(instructionsPath), 0o755))
		existing := copilotStartMarker + "\nold\n" + copilotEndMarker + "\n\n## After section\n"
		require.NoError(t, os.WriteFile(instructionsPath, []byte(existing), 0o644))

		a := NewCopilotAdapter()
		require.NoError(t, a.ApplyProject(context.Background(), domain.Project{Path: projectDir}, skills))

		content, _ := os.ReadFile(instructionsPath)
		assert.Contains(t, string(content), "## After section")
	})

	t.Run("appends block when no markers exist yet", func(t *testing.T) {
		t.Parallel()

		projectDir := t.TempDir()
		instructionsPath := filepath.Join(projectDir, copilotFilePath)
		require.NoError(t, os.MkdirAll(filepath.Dir(instructionsPath), 0o755))
		require.NoError(t, os.WriteFile(instructionsPath, []byte("# Existing content\n"), 0o644))

		a := NewCopilotAdapter()
		require.NoError(t, a.ApplyProject(context.Background(), domain.Project{Path: projectDir}, skills))

		content, _ := os.ReadFile(instructionsPath)
		s := string(content)
		assert.Contains(t, s, "# Existing content")
		assert.Contains(t, s, copilotStartMarker)
	})

	t.Run("empty skills produces empty managed block", func(t *testing.T) {
		t.Parallel()

		projectDir := t.TempDir()
		a := NewCopilotAdapter()
		require.NoError(t, a.ApplyProject(context.Background(), domain.Project{Path: projectDir}, []domain.Skill{}))

		content, _ := os.ReadFile(filepath.Join(projectDir, copilotFilePath))
		s := string(content)
		assert.Contains(t, s, copilotStartMarker)
		assert.Contains(t, s, copilotEndMarker)
		assert.NotContains(t, s, "Active Skills")
	})
}

func TestBuildManagedBlock(t *testing.T) {
	t.Parallel()

	t.Run("renders skill name and description", func(t *testing.T) {
		t.Parallel()
		skills := []domain.Skill{{Name: "alpha", Description: "Alpha skill"}}
		block := buildManagedBlock(skills)
		assert.Contains(t, block, "**alpha**")
		assert.Contains(t, block, "Alpha skill")
		assert.True(t, strings.HasPrefix(block, copilotStartMarker))
		assert.True(t, strings.HasSuffix(block, copilotEndMarker))
	})

	t.Run("empty skills produces marker-only block", func(t *testing.T) {
		t.Parallel()
		block := buildManagedBlock(nil)
		assert.Equal(t, copilotStartMarker+"\n"+copilotEndMarker, block)
	})
}
