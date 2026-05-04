package usecase

import "time"

// SkillProjectRef is a lightweight project reference attached to an aggregated skill.
type SkillProjectRef struct {
	ID        string
	Name      string
	Path      string // project root path (for display/navigation)
	SkillPath string // skill directory path (for reading SKILL.md)
}

// AggregatedSkill groups all locations (global + projects) for a single skill name.
type AggregatedSkill struct {
	Name        string
	Description string
	IsGlobal    bool
	GlobalPath  string // non-empty when IsGlobal
	Projects    []SkillProjectRef
	UpdatedAt   time.Time
}
