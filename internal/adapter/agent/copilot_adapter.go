package agent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

// ApplyProject regenerates the managed block inside copilot-instructions.md,
// preserving any user content outside the markers.
func (a *CopilotAdapter) ApplyProject(_ context.Context, project domain.Project, activeSkills []domain.Skill) error {
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
