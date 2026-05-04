package binding

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"skill-manager/internal/usecase"
)

// SkillsBinding exposes skill-related operations to the Wails frontend.
type SkillsBinding struct {
	list        *usecase.ListSkills
	listProject *usecase.ListProjectSkills
	listAll     *usecase.ListAllSkills
	copySkill   *usecase.CopySkill
	deleteSkill *usecase.DeleteSkill
	trashDir    string
}

func NewSkillsBinding(
	list *usecase.ListSkills,
	listProject *usecase.ListProjectSkills,
	listAll *usecase.ListAllSkills,
	copySkill *usecase.CopySkill,
	deleteSkill *usecase.DeleteSkill,
) *SkillsBinding {
	home, _ := os.UserHomeDir()
	return &SkillsBinding{
		list:        list,
		listProject: listProject,
		listAll:     listAll,
		copySkill:   copySkill,
		deleteSkill: deleteSkill,
		trashDir:    filepath.Join(home, ".skills-manager", "trash"),
	}
}

// List returns all global skills.
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

// ListByProject returns all skills discovered inside the given project directory.
func (b *SkillsBinding) ListByProject(ctx context.Context, projectID string) ([]SkillDTO, error) {
	skills, err := b.listProject.Execute(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("skills binding: list by project: %w", err)
	}
	dtos := make([]SkillDTO, len(skills))
	for i, s := range skills {
		dtos[i] = toSkillDTO(s)
	}
	return dtos, nil
}

// ListAll returns every skill aggregated and deduplicated via the SQLite cache.
func (b *SkillsBinding) ListAll(ctx context.Context) ([]AggregatedSkillDTO, error) {
	skills, err := b.listAll.Execute(ctx)
	if err != nil {
		return nil, fmt.Errorf("skills binding: list all: %w", err)
	}
	dtos := make([]AggregatedSkillDTO, len(skills))
	for i, s := range skills {
		dtos[i] = toAggregatedSkillDTO(s)
	}
	return dtos, nil
}

// CopySkill copies a skill directory into the target project's agent skills folder.
func (b *SkillsBinding) CopySkill(ctx context.Context, req CopySkillRequestDTO) error {
	return b.copySkill.Execute(ctx, usecase.CopySkillRequest{
		SkillID:         req.SkillID,
		SourceProjectID: req.SourceProjectID,
		TargetProjectID: req.TargetProjectID,
		Agent:           req.Agent,
	})
}

// DeleteSkill backs up the skill to trash then removes it from disk.
func (b *SkillsBinding) DeleteSkill(ctx context.Context, req DeleteSkillRequestDTO) error {
	return b.deleteSkill.Execute(ctx, usecase.DeleteSkillRequest{
		SkillID:   req.SkillID,
		ProjectID: req.ProjectID,
		TrashDir:  b.trashDir,
	})
}
