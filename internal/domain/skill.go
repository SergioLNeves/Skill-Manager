package domain

import "time"

// Skill represents a skill stored in the central repository.
type Skill struct {
	ID          string
	Name        string
	Path        string
	Description string
	UpdatedAt   time.Time
}
