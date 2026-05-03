package agent

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"skill-manager/internal/adapter/filesystem"
	"skill-manager/internal/domain"
)

// ClaudeAdapter activates skills by creating symlinks inside .claude/skills/ directories.
// Global path: ~/.claude/skills/<skill-name> → <skill-path>
// Project path: <project>/.claude/skills/<skill-name> → <skill-path>
type ClaudeAdapter struct {
	homeDir string
	symlink *filesystem.SymlinkManager
}

// NewClaudeAdapter creates a ClaudeAdapter.
// homeDir should be the user's home directory (os.UserHomeDir()).
func NewClaudeAdapter(homeDir string) *ClaudeAdapter {
	return &ClaudeAdapter{
		homeDir: homeDir,
		symlink: filesystem.NewSymlinkManager(""),
	}
}

func (a *ClaudeAdapter) Agent() domain.Agent {
	return domain.AgentClaude
}

// CapabilityCheck verifies that the global Claude skills directory is accessible.
func (a *ClaudeAdapter) CapabilityCheck(_ context.Context) error {
	dir := a.globalSkillsDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("claude adapter: cannot access global skills dir %s: %w", dir, err)
	}
	return nil
}

// ApplyGlobal syncs the global ~/.claude/skills/ directory to match activeSkills.
// Creates missing links, removes links for skills no longer active.
func (a *ClaudeAdapter) ApplyGlobal(ctx context.Context, activeSkills []domain.Skill) error {
	dir := a.globalSkillsDir()
	return a.sync(ctx, dir, activeSkills)
}

// ApplyProject syncs <project>/.claude/skills/ to match activeSkills.
func (a *ClaudeAdapter) ApplyProject(ctx context.Context, project domain.Project, activeSkills []domain.Skill) error {
	dir := filepath.Join(project.Path, ".claude", "skills")
	return a.sync(ctx, dir, activeSkills)
}

// sync brings dir into the desired state: one symlink per active skill, nothing extra.
func (a *ClaudeAdapter) sync(ctx context.Context, dir string, activeSkills []domain.Skill) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("claude adapter: mkdir %s: %w", dir, err)
	}

	// Build a set of desired link names.
	desired := make(map[string]domain.Skill, len(activeSkills))
	for _, s := range activeSkills {
		desired[s.Name] = s
	}

	// Walk existing entries and remove stale managed links.
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("claude adapter: readdir %s: %w", dir, err)
	}

	for _, e := range entries {
		if e.Type()&os.ModeSymlink == 0 {
			// Not a symlink — leave it alone (user-managed file or dir).
			continue
		}
		if _, ok := desired[e.Name()]; !ok {
			linkPath := filepath.Join(dir, e.Name())
			if removeErr := os.Remove(linkPath); removeErr != nil {
				return fmt.Errorf("claude adapter: remove stale link %s: %w", linkPath, removeErr)
			}
		}
	}

	// Ensure all desired links exist.
	for name, skill := range desired {
		linkPath := filepath.Join(dir, name)
		skillMgr := filesystem.NewSymlinkManager(filepath.Dir(skill.Path))
		if ensureErr := skillMgr.EnsureLink(ctx, skill.Path, linkPath); ensureErr != nil {
			if errors.Is(ensureErr, filesystem.ErrTargetNotManaged) {
				return fmt.Errorf("claude adapter: link %s is occupied by an unmanaged path: %w", linkPath, ensureErr)
			}
			return fmt.Errorf("claude adapter: ensure link %s: %w", linkPath, ensureErr)
		}
	}

	return nil
}

func (a *ClaudeAdapter) globalSkillsDir() string {
	return filepath.Join(a.homeDir, ".claude", "skills")
}
