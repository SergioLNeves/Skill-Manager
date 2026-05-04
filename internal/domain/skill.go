package domain

import "time"

// SkillSource indicates where a skill was discovered.
type SkillSource string

const (
	SkillSourceGlobal  SkillSource = "global"
	SkillSourceProject SkillSource = "project"
)

// Skill represents a skill discovered from a source directory.
type Skill struct {
	ID             string
	Name           string
	Path           string
	Description    string
	Source         SkillSource
	OwnerProjectID string // non-empty only when Source == SkillSourceProject
	UpdatedAt      time.Time
}
