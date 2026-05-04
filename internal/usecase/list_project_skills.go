package usecase

import (
	"context"
	"errors"
	"fmt"

	"skill-manager/internal/domain"
)

// ListProjectSkills returns all skills discovered inside a project directory.
type ListProjectSkills struct {
	projects     ProjectRepository
	projectSkills ProjectSkillRepository
}

func NewListProjectSkills(projects ProjectRepository, projectSkills ProjectSkillRepository) *ListProjectSkills {
	return &ListProjectSkills{projects: projects, projectSkills: projectSkills}
}

func (uc *ListProjectSkills) Execute(ctx context.Context, projectID string) ([]domain.Skill, error) {
	project, err := uc.projects.GetByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, domain.ErrProjectNotFound) {
			return nil, domain.ErrProjectNotFound
		}
		return nil, fmt.Errorf("list project skills: fetch project: %w", err)
	}
	skills, err := uc.projectSkills.ListByProject(ctx, project)
	if err != nil {
		return nil, fmt.Errorf("list project skills: scan: %w", err)
	}
	return skills, nil
}
