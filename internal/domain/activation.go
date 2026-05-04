package domain

import "time"

// Scope determines whether an activation is global or project-scoped.
type Scope string

const (
	ScopeGlobal  Scope = "global"
	ScopeProject Scope = "project"
)

// Activation records that a skill is active for a given agent and scope.
type Activation struct {
	ID        int64
	SkillID   string
	Agent     Agent
	Scope     Scope
	ProjectID *string // nil when Scope == ScopeGlobal
	AppliedAt time.Time
}
