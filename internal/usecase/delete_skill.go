package usecase

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// DeleteSkillRequest carries the parameters for a skill deletion.
type DeleteSkillRequest struct {
	SkillID   string
	ProjectID string // empty when skill is global
	TrashDir  string // e.g. ~/.skills-manager/trash
}

// DeleteSkill backs up the skill directory to TrashDir then removes it from disk.
// The activation records referencing the skill are also cleaned up.
type DeleteSkill struct {
	skills        SkillRepository
	projects      ProjectRepository
	projectSkills ProjectSkillRepository
	activations   ActivationRepository
}

func NewDeleteSkill(skills SkillRepository, projects ProjectRepository, projectSkills ProjectSkillRepository, activations ActivationRepository) *DeleteSkill {
	return &DeleteSkill{
		skills:        skills,
		projects:      projects,
		projectSkills: projectSkills,
		activations:   activations,
	}
}

func (uc *DeleteSkill) Execute(ctx context.Context, req DeleteSkillRequest) error {
	skillPath, err := resolveSkillPath(ctx, req.SkillID, req.ProjectID, uc.skills, uc.projects, uc.projectSkills)
	if err != nil {
		return fmt.Errorf("delete skill: resolve: %w", err)
	}

	skillDir := filepath.Dir(skillPath)
	skillName := filepath.Base(skillDir)
	timestamp := time.Now().UTC().Format("20060102T150405Z")
	backupDst := filepath.Join(req.TrashDir, fmt.Sprintf("%s-%s", skillName, timestamp))

	if err := os.MkdirAll(req.TrashDir, 0o755); err != nil {
		return fmt.Errorf("delete skill: create trash dir: %w", err)
	}
	if err := copyDir(skillDir, backupDst); err != nil {
		return fmt.Errorf("delete skill: backup to trash: %w", err)
	}

	if err := os.RemoveAll(skillDir); err != nil {
		return fmt.Errorf("delete skill: remove physical: %w", err)
	}

	// Best-effort: clean up activation records.
	activations, _ := uc.activations.List(ctx, ActivationFilter{SkillID: req.SkillID})
	for _, a := range activations {
		_ = uc.activations.Delete(ctx, a.ID)
	}

	return nil
}
