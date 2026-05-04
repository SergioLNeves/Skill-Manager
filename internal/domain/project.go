package domain

import "time"

// Project is a registered codebase that can consume skills.
type Project struct {
	ID             string
	Name           string
	Path           string
	DetectedAgents []Agent
	AddedAt        time.Time
}

// SupportsAgent reports whether the project has the given agent configured.
func (p Project) SupportsAgent(a Agent) bool {
	for _, detected := range p.DetectedAgents {
		if detected == a {
			return true
		}
	}
	return false
}
