package domain

import "time"

// SkillSource indicates where a skill was discovered.
type SkillSource string

const (
	SkillSourceGitHub  SkillSource = "github"
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

	// GitHub-sourced fields (non-empty when Source == SkillSourceGitHub)
	Repo    string // "owner/repo"
	Ref     string // resolved commit SHA
	SubPath string // path of skill dir relative to repo root
}
