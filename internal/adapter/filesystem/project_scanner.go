package filesystem

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"

	"skill-manager/internal/domain"
	"skill-manager/internal/usecase"
)

const maxScanDepth = 3

// excludedDirs are directory names skipped during scan.
var excludedDirs = map[string]bool{
	"node_modules": true,
	"vendor":       true,
	".git":         true,
	"dist":         true,
	"build":        true,
	".cache":       true,
}

// ProjectScanner walks workspace roots and detects project directories.
type ProjectScanner struct {
	fs afero.Fs
}

// NewProjectScanner creates a scanner backed by the given filesystem.
func NewProjectScanner(fs afero.Fs) *ProjectScanner {
	return &ProjectScanner{fs: fs}
}

// Scan returns project candidates found under the given roots.
func (s *ProjectScanner) Scan(_ context.Context, roots []string) ([]usecase.ProjectCandidate, error) {
	var candidates []usecase.ProjectCandidate
	seen := map[string]bool{}

	for _, root := range roots {
		found, err := s.scanRoot(root, seen)
		if err != nil {
			return nil, fmt.Errorf("project scanner: scan %s: %w", root, err)
		}
		candidates = append(candidates, found...)
	}
	return candidates, nil
}

func (s *ProjectScanner) scanRoot(root string, seen map[string]bool) ([]usecase.ProjectCandidate, error) {
	var candidates []usecase.ProjectCandidate
	err := s.walk(root, 0, &candidates, seen)
	return candidates, err
}

func (s *ProjectScanner) walk(dir string, depth int, out *[]usecase.ProjectCandidate, seen map[string]bool) error {
	if depth > maxScanDepth {
		return nil
	}

	entries, err := afero.ReadDir(s.fs, dir)
	if err != nil {
		return nil // skip unreadable directories silently
	}

	isProject := false
	for _, e := range entries {
		if e.IsDir() && e.Name() == ".git" {
			isProject = true
			break
		}
	}

	if isProject {
		abs := dir
		if !seen[abs] {
			seen[abs] = true
			*out = append(*out, usecase.ProjectCandidate{
				Name:           filepath.Base(dir),
				Path:           abs,
				DetectedAgents: detectAgents(s.fs, dir),
			})
		}
		return nil // don't recurse into git projects
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasPrefix(name, ".") || excludedDirs[name] {
			continue
		}
		if err := s.walk(filepath.Join(dir, name), depth+1, out, seen); err != nil {
			return err
		}
	}
	return nil
}

// detectAgents inspects the project directory for known agent configuration files.
func detectAgents(fs afero.Fs, dir string) []domain.Agent {
	var agents []domain.Agent

	if hasClaude(fs, dir) {
		agents = append(agents, domain.AgentClaude)
	}
	if hasCopilot(fs, dir) {
		agents = append(agents, domain.AgentCopilot)
	}
	return agents
}

func hasClaude(fs afero.Fs, dir string) bool {
	return existsInFS(fs, filepath.Join(dir, ".claude")) ||
		existsInFS(fs, filepath.Join(dir, "CLAUDE.md")) ||
		existsInFS(fs, filepath.Join(dir, "AGENTS.md"))
}

func hasCopilot(fs afero.Fs, dir string) bool {
	return existsInFS(fs, filepath.Join(dir, ".github", "copilot-instructions.md")) ||
		existsInFS(fs, filepath.Join(dir, ".github", "instructions"))
}

func existsInFS(fs afero.Fs, path string) bool {
	_, err := fs.Stat(path)
	return err == nil
}
