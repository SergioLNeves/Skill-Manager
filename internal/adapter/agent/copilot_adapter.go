package agent

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"skill-manager/internal/adapter/filesystem"
	"skill-manager/internal/domain"
)

const (
	copilotStartMarker = "<!-- skills-manager:start -->"
	copilotEndMarker   = "<!-- skills-manager:end -->"
	copilotFilePath    = ".github/copilot-instructions.md"
)

// CopilotAdapter activates skills by injecting a managed block into
// <project>/.github/copilot-instructions.md.
// Global activation is not supported and is silently skipped.
type CopilotAdapter struct{}

func NewCopilotAdapter() *CopilotAdapter {
	return &CopilotAdapter{}
}

func (a *CopilotAdapter) Agent() domain.Agent {
	return domain.AgentCopilot
}

// CapabilityCheck always succeeds — no global state required.
func (a *CopilotAdapter) CapabilityCheck(_ context.Context) error {
	return nil
}

// ApplyGlobal is a no-op for Copilot (no global instructions file concept).
func (a *CopilotAdapter) ApplyGlobal(_ context.Context, _ []domain.Skill) error {
	return nil
}

// ApplyProject syncs <project>/.github/skills/ to match activeSkills (symlinks)
// and regenerates the managed block inside copilot-instructions.md.
func (a *CopilotAdapter) ApplyProject(ctx context.Context, project domain.Project, activeSkills []domain.Skill) error {
	if err := a.syncSkillsDir(ctx, project.Path, activeSkills); err != nil {
		return err
	}

	instructionsPath := filepath.Join(project.Path, copilotFilePath)

	existing, err := os.ReadFile(instructionsPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("copilot adapter: read %s: %w", instructionsPath, err)
	}

	updated := buildInstructions(string(existing), activeSkills)

	if err := os.MkdirAll(filepath.Dir(instructionsPath), 0o755); err != nil {
		return fmt.Errorf("copilot adapter: mkdir for %s: %w", instructionsPath, err)
	}
	if err := os.WriteFile(instructionsPath, []byte(updated), 0o644); err != nil {
		return fmt.Errorf("copilot adapter: write %s: %w", instructionsPath, err)
	}
	return nil
}

// syncSkillsDir syncs <projectPath>/.github/skills/ to exactly the active skills,
// creating symlinks for new skills and removing stale ones.
func (a *CopilotAdapter) syncSkillsDir(ctx context.Context, projectPath string, activeSkills []domain.Skill) error {
	dir := filepath.Join(projectPath, ".github", "skills")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("copilot adapter: mkdir %s: %w", dir, err)
	}

	desired := make(map[string]domain.Skill, len(activeSkills))
	for _, s := range activeSkills {
		desired[s.Name] = s
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("copilot adapter: readdir %s: %w", dir, err)
	}
	for _, e := range entries {
		if e.Type()&os.ModeSymlink == 0 {
			continue
		}
		if _, ok := desired[e.Name()]; !ok {
			linkPath := filepath.Join(dir, e.Name())
			if removeErr := os.Remove(linkPath); removeErr != nil {
				return fmt.Errorf("copilot adapter: remove stale link %s: %w", linkPath, removeErr)
			}
		}
	}

	for name, skill := range desired {
		linkPath := filepath.Join(dir, name)
		skillMgr := filesystem.NewSymlinkManager(filepath.Dir(skill.Path))
		if ensureErr := skillMgr.EnsureLink(ctx, skill.Path, linkPath); ensureErr != nil {
			if errors.Is(ensureErr, filesystem.ErrTargetNotManaged) {
				return fmt.Errorf("copilot adapter: link %s is occupied by an unmanaged path: %w", linkPath, ensureErr)
			}
			return fmt.Errorf("copilot adapter: ensure link %s: %w", linkPath, ensureErr)
		}
	}
	return nil
}

// buildInstructions splices the managed block into content, preserving user sections.
func buildInstructions(existing string, skills []domain.Skill) string {
	block := buildManagedBlock(skills)

	startIdx := strings.Index(existing, copilotStartMarker)
	endIdx := strings.Index(existing, copilotEndMarker)

	if startIdx != -1 && endIdx != -1 && endIdx > startIdx {
		// Replace the existing managed block (including markers).
		before := existing[:startIdx]
		after := existing[endIdx+len(copilotEndMarker):]
		return strings.TrimRight(before, "\n") + "\n\n" + block + strings.TrimLeft(after, "\n")
	}

	// No existing markers — append the block.
	if existing == "" {
		return block
	}
	return strings.TrimRight(existing, "\n") + "\n\n" + block
}

// buildManagedBlock renders the skills list between the managed markers.
func buildManagedBlock(skills []domain.Skill) string {
	if len(skills) == 0 {
		return copilotStartMarker + "\n" + copilotEndMarker
	}

	var sb strings.Builder
	sb.WriteString(copilotStartMarker)
	sb.WriteString("\n## Active Skills\n\n")
	for _, s := range skills {
		sb.WriteString(fmt.Sprintf("- **%s**: %s\n", s.Name, s.Description))
	}
	sb.WriteString(copilotEndMarker)
	return sb.String()
}
