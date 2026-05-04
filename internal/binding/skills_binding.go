package binding

import (
	"context"
	"fmt"

	"skill-manager/internal/usecase"
)

// SkillsBinding exposes skill-related operations to the Wails frontend.
type SkillsBinding struct {
	list *usecase.ListSkills
}

func NewSkillsBinding(list *usecase.ListSkills) *SkillsBinding {
	return &SkillsBinding{list: list}
}

// List returns all skills from the central repository.
func (b *SkillsBinding) List(ctx context.Context) ([]SkillDTO, error) {
	skills, err := b.list.Execute(ctx)
	if err != nil {
		return nil, fmt.Errorf("skills binding: %w", err)
	}
	dtos := make([]SkillDTO, len(skills))
	for i, s := range skills {
		dtos[i] = toSkillDTO(s)
	}
	return dtos, nil
}
