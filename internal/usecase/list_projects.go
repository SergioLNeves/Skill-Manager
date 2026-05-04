package usecase

import (
	"context"
	"fmt"

	"skill-manager/internal/domain"
)

// ListProjects retrieves all registered projects.
type ListProjects struct {
	projects ProjectRepository
}

func NewListProjects(projects ProjectRepository) *ListProjects {
	return &ListProjects{projects: projects}
}

func (uc *ListProjects) Execute(ctx context.Context) ([]domain.Project, error) {
	list, err := uc.projects.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	return list, nil
}
