package usecase

import (
	"context"
	"fmt"
)

// DeleteProject removes a registered project and its activations (via DB cascade).
type DeleteProject struct {
	projects ProjectRepository
}

func NewDeleteProject(projects ProjectRepository) *DeleteProject {
	return &DeleteProject{projects: projects}
}

func (uc *DeleteProject) Execute(ctx context.Context, id string) error {
	if err := uc.projects.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete project %s: %w", id, err)
	}
	return nil
}
