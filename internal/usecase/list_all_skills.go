package usecase

import (
	"context"
	"fmt"

	"skill-manager/internal/domain"
)

// SkillWithProject bundles a skill with its owning project name for display.
type SkillWithProject struct {
	domain.Skill
	OwnerProjectName string
}

// ListAllSkills aggregates global skills plus all project skills across every registered project.
type ListAllSkills struct {
	skills        SkillRepository
	projects      ProjectRepository
	projectSkills ProjectSkillRepository
}

func NewListAllSkills(skills SkillRepository, projects ProjectRepository, projectSkills ProjectSkillRepository) *ListAllSkills {
	return &ListAllSkills{skills: skills, projects: projects, projectSkills: projectSkills}
}

func (uc *ListAllSkills) Execute(ctx context.Context) ([]SkillWithProject, error) {
	var result []SkillWithProject

	globals, err := uc.skills.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list all skills: globals: %w", err)
	}
	for _, s := range globals {
		result = append(result, SkillWithProject{Skill: s})
	}

	projects, err := uc.projects.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list all skills: projects: %w", err)
	}
	for _, p := range projects {
		skills, err := uc.projectSkills.ListByProject(ctx, p)
		if err != nil {
			continue // skip unreachable project directories
		}
		for _, s := range skills {
			result = append(result, SkillWithProject{Skill: s, OwnerProjectName: p.Name})
		}
	}

	return result, nil
}
