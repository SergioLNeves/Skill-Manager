package agent

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"skill-manager/internal/domain"
)

func TestClaudeAdapter_Agent(t *testing.T) {
	a := NewClaudeAdapter(t.TempDir())
	assert.Equal(t, domain.AgentClaude, a.Agent())
}

func TestClaudeAdapter_CapabilityCheck(t *testing.T) {
	t.Parallel()

	t.Run("creates global skills dir when missing", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		a := NewClaudeAdapter(home)

		require.NoError(t, a.CapabilityCheck(context.Background()))
		_, err := os.Stat(filepath.Join(home, ".claude", "skills"))
		assert.NoError(t, err)
	})
}

func TestClaudeAdapter_ApplyGlobal(t *testing.T) {
	t.Parallel()

	t.Run("creates symlink for each active skill", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		skillRoot := t.TempDir()
		skill := makeSkill(t, skillRoot, "go-review")
		a := NewClaudeAdapter(home)

		err := a.ApplyGlobal(context.Background(), []domain.Skill{skill})

		require.NoError(t, err)
		link := filepath.Join(home, ".claude", "skills", "go-review")
		dest, readErr := os.Readlink(link)
		require.NoError(t, readErr)
		assert.Equal(t, skill.Path, dest)
	})

	t.Run("removes stale managed links", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		skillRoot := t.TempDir()
		skill1 := makeSkill(t, skillRoot, "alpha")
		skill2 := makeSkill(t, skillRoot, "beta")
		a := NewClaudeAdapter(home)

		// Activate both.
		require.NoError(t, a.ApplyGlobal(context.Background(), []domain.Skill{skill1, skill2}))
		// Now only alpha is active.
		require.NoError(t, a.ApplyGlobal(context.Background(), []domain.Skill{skill1}))

		betaLink := filepath.Join(home, ".claude", "skills", "beta")
		_, statErr := os.Lstat(betaLink)
		assert.True(t, os.IsNotExist(statErr))
	})

	t.Run("is idempotent", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		skillRoot := t.TempDir()
		skill := makeSkill(t, skillRoot, "my-skill")
		a := NewClaudeAdapter(home)

		require.NoError(t, a.ApplyGlobal(context.Background(), []domain.Skill{skill}))
		require.NoError(t, a.ApplyGlobal(context.Background(), []domain.Skill{skill}))

		link := filepath.Join(home, ".claude", "skills", "my-skill")
		_, readErr := os.Readlink(link)
		assert.NoError(t, readErr)
	})

	t.Run("empty active skills removes all managed links", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		skillRoot := t.TempDir()
		skill := makeSkill(t, skillRoot, "to-remove")
		a := NewClaudeAdapter(home)

		require.NoError(t, a.ApplyGlobal(context.Background(), []domain.Skill{skill}))
		require.NoError(t, a.ApplyGlobal(context.Background(), []domain.Skill{}))

		link := filepath.Join(home, ".claude", "skills", "to-remove")
		_, statErr := os.Lstat(link)
		assert.True(t, os.IsNotExist(statErr))
	})
}

func TestClaudeAdapter_ApplyProject(t *testing.T) {
	t.Parallel()

	t.Run("creates symlink inside project .claude/skills", func(t *testing.T) {
		t.Parallel()

		projectDir := t.TempDir()
		skillRoot := t.TempDir()
		skill := makeSkill(t, skillRoot, "debug-tips")
		project := domain.Project{Path: projectDir}
		a := NewClaudeAdapter(t.TempDir())

		err := a.ApplyProject(context.Background(), project, []domain.Skill{skill})

		require.NoError(t, err)
		link := filepath.Join(projectDir, ".claude", "skills", "debug-tips")
		dest, readErr := os.Readlink(link)
		require.NoError(t, readErr)
		assert.Equal(t, skill.Path, dest)
	})
}

// makeSkill creates a real skill directory and returns a domain.Skill for it.
func makeSkill(t *testing.T, root, name string) domain.Skill {
	t.Helper()
	path := filepath.Join(root, name)
	require.NoError(t, os.MkdirAll(path, 0o755))
	return domain.Skill{Name: name, Path: path, Description: name}
}
