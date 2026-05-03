package filesystem

import (
	"context"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"skill-manager/internal/domain"
)

func TestProjectScanner_Scan(t *testing.T) {
	t.Parallel()

	t.Run("detects project with .git directory", func(t *testing.T) {
		t.Parallel()

		fs := makeProjectFS(t, "/workspace/my-service", nil)
		scanner := NewProjectScanner(fs)

		candidates, err := scanner.Scan(context.Background(), []string{"/workspace"})

		require.NoError(t, err)
		require.Len(t, candidates, 1)
		assert.Equal(t, "my-service", candidates[0].Name)
		assert.Equal(t, "/workspace/my-service", candidates[0].Path)
	})

	t.Run("detects Claude agent via .claude directory", func(t *testing.T) {
		t.Parallel()

		fs := makeProjectFS(t, "/ws/svc", []agentMarker{
			{path: ".claude", isDir: true},
		})
		scanner := NewProjectScanner(fs)

		candidates, err := scanner.Scan(context.Background(), []string{"/ws"})

		require.NoError(t, err)
		require.Len(t, candidates, 1)
		assert.Contains(t, candidates[0].DetectedAgents, domain.AgentClaude)
	})

	t.Run("detects Claude agent via CLAUDE.md file", func(t *testing.T) {
		t.Parallel()

		fs := makeProjectFS(t, "/ws/svc", []agentMarker{
			{path: "CLAUDE.md", content: "# Claude"},
		})
		scanner := NewProjectScanner(fs)

		candidates, err := scanner.Scan(context.Background(), []string{"/ws"})

		require.NoError(t, err)
		assert.Contains(t, candidates[0].DetectedAgents, domain.AgentClaude)
	})

	t.Run("detects Claude agent via AGENTS.md file", func(t *testing.T) {
		t.Parallel()

		fs := makeProjectFS(t, "/ws/svc", []agentMarker{
			{path: "AGENTS.md", content: "# Agents"},
		})
		scanner := NewProjectScanner(fs)

		candidates, err := scanner.Scan(context.Background(), []string{"/ws"})

		require.NoError(t, err)
		assert.Contains(t, candidates[0].DetectedAgents, domain.AgentClaude)
	})

	t.Run("detects Copilot agent via copilot-instructions.md", func(t *testing.T) {
		t.Parallel()

		fs := makeProjectFS(t, "/ws/svc", []agentMarker{
			{path: ".github/copilot-instructions.md", content: "# Copilot"},
		})
		scanner := NewProjectScanner(fs)

		candidates, err := scanner.Scan(context.Background(), []string{"/ws"})

		require.NoError(t, err)
		assert.Contains(t, candidates[0].DetectedAgents, domain.AgentCopilot)
	})

	t.Run("detects both agents when both are configured", func(t *testing.T) {
		t.Parallel()

		fs := makeProjectFS(t, "/ws/svc", []agentMarker{
			{path: "CLAUDE.md", content: "# Claude"},
			{path: ".github/copilot-instructions.md", content: "# Copilot"},
		})
		scanner := NewProjectScanner(fs)

		candidates, err := scanner.Scan(context.Background(), []string{"/ws"})

		require.NoError(t, err)
		require.Len(t, candidates, 1)
		assert.Len(t, candidates[0].DetectedAgents, 2)
	})

	t.Run("skips excluded directories", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		// node_modules with a .git inside — should be skipped
		makeProjectFS(t, "/ws/node_modules/leaked-project", nil)
		require.NoError(t, fs.MkdirAll("/ws/node_modules/.git", 0o755))
		// real project at root level
		makeProjectFSOn(t, fs, "/ws/real-project", nil)

		scanner := NewProjectScanner(fs)
		candidates, err := scanner.Scan(context.Background(), []string{"/ws"})

		require.NoError(t, err)
		require.Len(t, candidates, 1)
		assert.Equal(t, "real-project", candidates[0].Name)
	})

	t.Run("skips hidden directories", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		makeProjectFSOn(t, fs, "/ws/.hidden-project", nil)
		makeProjectFSOn(t, fs, "/ws/visible-project", nil)

		scanner := NewProjectScanner(fs)
		candidates, err := scanner.Scan(context.Background(), []string{"/ws"})

		require.NoError(t, err)
		require.Len(t, candidates, 1)
		assert.Equal(t, "visible-project", candidates[0].Name)
	})

	t.Run("does not recurse into git project subdirectories", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		makeProjectFSOn(t, fs, "/ws/parent", nil)
		// nested repo inside parent — should be ignored
		makeProjectFSOn(t, fs, "/ws/parent/nested", nil)

		scanner := NewProjectScanner(fs)
		candidates, err := scanner.Scan(context.Background(), []string{"/ws"})

		require.NoError(t, err)
		assert.Len(t, candidates, 1)
		assert.Equal(t, "parent", candidates[0].Name)
	})

	t.Run("deduplicates when same root is given twice", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		makeProjectFSOn(t, fs, "/ws/svc", nil)

		scanner := NewProjectScanner(fs)
		candidates, err := scanner.Scan(context.Background(), []string{"/ws", "/ws"})

		require.NoError(t, err)
		assert.Len(t, candidates, 1)
	})

	t.Run("returns empty slice when no projects found", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		require.NoError(t, fs.MkdirAll("/ws/not-a-repo", 0o755))

		scanner := NewProjectScanner(fs)
		candidates, err := scanner.Scan(context.Background(), []string{"/ws"})

		require.NoError(t, err)
		assert.Empty(t, candidates)
	})

	t.Run("respects max depth limit", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		// create a project 4 levels deep (exceeds maxScanDepth=3)
		deep := "/ws/a/b/c/d/too-deep"
		makeProjectFSOn(t, fs, deep, nil)
		// one within limit
		makeProjectFSOn(t, fs, "/ws/a/b/within-limit", nil)

		scanner := NewProjectScanner(fs)
		candidates, err := scanner.Scan(context.Background(), []string{"/ws"})

		require.NoError(t, err)
		for _, c := range candidates {
			assert.NotEqual(t, "too-deep", c.Name)
		}
	})
}

// agentMarker describes a file or directory to create inside a project.
type agentMarker struct {
	path    string
	content string
	isDir   bool
}

// makeProjectFS creates an in-memory fs with a project at dir.
func makeProjectFS(t *testing.T, dir string, markers []agentMarker) afero.Fs {
	t.Helper()
	fs := afero.NewMemMapFs()
	makeProjectFSOn(t, fs, dir, markers)
	return fs
}

// makeProjectFSOn sets up a project at dir on an existing fs.
func makeProjectFSOn(t *testing.T, fs afero.Fs, dir string, markers []agentMarker) {
	t.Helper()
	require.NoError(t, fs.MkdirAll(dir+"/.git", 0o755))
	for _, m := range markers {
		full := dir + "/" + m.path
		if m.isDir {
			require.NoError(t, fs.MkdirAll(full, 0o755))
		} else {
			require.NoError(t, fs.MkdirAll(dir+"/"+dirOf(m.path), 0o755))
			require.NoError(t, afero.WriteFile(fs, full, []byte(m.content), 0o644))
		}
	}
}

func dirOf(p string) string {
	idx := len(p) - len(p[len(p)-1:])
	for idx > 0 && p[idx-1] != '/' {
		idx--
	}
	if idx == 0 {
		return "."
	}
	return p[:idx-1]
}
