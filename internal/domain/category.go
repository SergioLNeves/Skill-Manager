package domain

import "time"

// Category groups skills by user-defined topic.
type Category struct {
	ID          int64
	Name        string
	Description string
	CreatedAt   time.Time
}

// ProjectCategoryLink represents a persisted association between a project,
// a category and an agent. While the link exists, every skill belonging to
// the category is kept copied into the project's agent skills folder.
type ProjectCategoryLink struct {
	ProjectID  string
	CategoryID int64
	Agent      Agent
	Category   Category
}
