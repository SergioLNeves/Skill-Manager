package filesystem

import (
	"context"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"skill-manager/internal/domain"
)

func TestSkillRepository_List(t *testing.T) {
	t.Parallel()

	t.Run("returns skills with parsed frontmatter", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		writeSkillFile(t, fs, "/skills/go-review", `---
name: go-review
description: Go code review checklist
---

# Content here`)

		repo := NewSkillRepository(fs, "/skills")
		skills, err := repo.List(context.Background())

		require.NoError(t, err)
		require.Len(t, skills, 1)
		assert.Equal(t, "go-review", skills[0].Name)
		assert.Equal(t, "Go code review checklist", skills[0].Description)
		assert.NotEmpty(t, skills[0].ID)
		assert.Equal(t, "/skills/go-review", skills[0].Path)
	})

	t.Run("returns multiple skills sorted by filesystem order", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		writeSkillFile(t, fs, "/skills/alpha", "# Alpha")
		writeSkillFile(t, fs, "/skills/beta", "# Beta")

		repo := NewSkillRepository(fs, "/skills")
		skills, err := repo.List(context.Background())

		require.NoError(t, err)
		assert.Len(t, skills, 2)
	})

	t.Run("skips directories starting with dot", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		writeSkillFile(t, fs, "/skills/.hidden", "# hidden")
		writeSkillFile(t, fs, "/skills/visible", "# visible")

		repo := NewSkillRepository(fs, "/skills")
		skills, err := repo.List(context.Background())

		require.NoError(t, err)
		assert.Len(t, skills, 1)
		assert.Equal(t, "visible", skills[0].Name)
	})

	t.Run("skips directories without SKILL.md", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		require.NoError(t, fs.MkdirAll("/skills/empty-dir", 0o755))
		writeSkillFile(t, fs, "/skills/with-skill", "# skill")

		repo := NewSkillRepository(fs, "/skills")
		skills, err := repo.List(context.Background())

		require.NoError(t, err)
		assert.Len(t, skills, 1)
	})

	t.Run("returns error when root does not exist", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		repo := NewSkillRepository(fs, "/nonexistent")

		_, err := repo.List(context.Background())
		require.Error(t, err)
	})

	t.Run("uses name as description when description field is absent", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		writeSkillFile(t, fs, "/skills/my-skill", `---
name: my-skill
---
`)
		repo := NewSkillRepository(fs, "/skills")
		skills, err := repo.List(context.Background())

		require.NoError(t, err)
		require.Len(t, skills, 1)
		assert.Equal(t, "my-skill", skills[0].Description)
	})

	t.Run("ID is stable across calls for the same skill name", func(t *testing.T) {
		t.Parallel()

		id1 := stableID("go-review")
		id2 := stableID("go-review")
		assert.Equal(t, id1, id2)

		idOther := stableID("pt-review")
		assert.NotEqual(t, id1, idOther)
	})
}

func TestSkillRepository_GetByID(t *testing.T) {
	t.Parallel()

	t.Run("returns skill when id matches", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		writeSkillFile(t, fs, "/skills/go-review", `---
description: Go code review
---`)
		repo := NewSkillRepository(fs, "/skills")
		id := stableID("go-review")

		skill, err := repo.GetByID(context.Background(), id)

		require.NoError(t, err)
		assert.Equal(t, "go-review", skill.Name)
	})

	t.Run("returns ErrSkillNotFound for unknown id", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		writeSkillFile(t, fs, "/skills/go-review", "# Go")
		repo := NewSkillRepository(fs, "/skills")

		_, err := repo.GetByID(context.Background(), "nonexistent-id")

		require.ErrorIs(t, err, domain.ErrSkillNotFound)
	})
}

func TestParseFrontmatter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		want    map[string]string
	}{
		{
			name:    "parses name and description",
			content: "---\nname: my-skill\ndescription: Does things\n---\n# Body",
			want:    map[string]string{"name": "my-skill", "description": "Does things"},
		},
		{
			name:    "returns empty map when no frontmatter",
			content: "# Just markdown",
			want:    map[string]string{},
		},
		{
			name:    "returns empty map when opening delimiter missing",
			content: "name: foo\n---",
			want:    map[string]string{},
		},
		{
			name:    "returns empty map when closing delimiter missing",
			content: "---\nname: foo",
			want:    map[string]string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := parseFrontmatter([]byte(tc.content))
			assert.Equal(t, tc.want, got)
		})
	}
}

// writeSkillFile creates a SKILL.md inside the given directory path.
func writeSkillFile(t *testing.T, fs afero.Fs, dir, content string) {
	t.Helper()
	require.NoError(t, fs.MkdirAll(dir, 0o755))
	require.NoError(t, afero.WriteFile(fs, dir+"/SKILL.md", []byte(content), 0o644))
}
