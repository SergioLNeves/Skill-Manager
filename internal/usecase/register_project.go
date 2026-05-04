package usecase

import (
	"context"
	"fmt"

	"skill-manager/internal/domain"
)

// RegisterProject persists a manually added project.
type RegisterProject struct {
	projects ProjectRepository
}

func NewRegisterProject(projects ProjectRepository) *RegisterProject {
	return &RegisterProject{projects: projects}
}

func (uc *RegisterProject) Execute(ctx context.Context, p domain.Project) error {
	if err := uc.projects.Save(ctx, p); err != nil {
		return fmt.Errorf("register project %q: %w", p.Path, err)
	}
	return nil
}
