package usecase

import (
	"context"
	"fmt"

	"skill-manager/internal/domain"
)

// ListSkills retrieves all skills from the central repository.
type ListSkills struct {
	skills SkillRepository
}

func NewListSkills(skills SkillRepository) *ListSkills {
	return &ListSkills{skills: skills}
}

func (uc *ListSkills) Execute(ctx context.Context) ([]domain.Skill, error) {
	list, err := uc.skills.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list skills: %w", err)
	}
	return list, nil
}
