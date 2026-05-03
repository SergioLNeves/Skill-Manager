package domain

// Conflict signals that a skill is already active globally and the caller
// is trying to activate it for a specific project under the same agent.
type Conflict struct {
	SkillID           string
	Agent             Agent
	GlobalActivation  *Activation
	ProjectActivation *Activation
}

// ConflictResolution is the user's chosen way to resolve a Conflict.
type ConflictResolution int

const (
	// ResolutionKeepGlobal cancels the project activation and keeps only global.
	ResolutionKeepGlobal ConflictResolution = iota + 1
	// ResolutionOverrideProject keeps both records but marks the project one as
	// an explicit override that shadows the global for that project.
	ResolutionOverrideProject
)
