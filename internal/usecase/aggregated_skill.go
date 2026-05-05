package usecase

import (
	"fmt"
	"time"
)

// CategorySkillPath is a skill name paired with its resolved source path.
// Used internally when copying category skills to a project.
type CategorySkillPath struct {
	SkillName string
	Path      string
}

// ResolveCategorySkillPaths resolves a skill to its preferred filesystem path.
// Global path is preferred; falls back to first project path.
func (s AggregatedSkill) PreferredPath() (string, error) {
	if s.IsGlobal && s.GlobalPath != "" {
		return s.GlobalPath, nil
	}
	if len(s.Projects) > 0 {
		return s.Projects[0].SkillPath, nil
	}
	return "", fmt.Errorf("no path available for skill %q", s.Name)
}

// SkillProjectRef is a lightweight project reference attached to an aggregated skill.
type SkillProjectRef struct {
	ID        string
	Name      string
	Path      string // project root path (for display/navigation)
	SkillPath string // skill directory path (for reading SKILL.md)
}

// AggregatedSkill groups all locations (global + projects) for a single skill name.
type AggregatedSkill struct {
	Name         string
	Description  string
	CategoryID   *int64 // nil when unassigned
	CategoryName string // empty when unassigned
	IsGlobal     bool
	GlobalPath   string // non-empty when IsGlobal
	Projects     []SkillProjectRef
	UpdatedAt    time.Time
}
