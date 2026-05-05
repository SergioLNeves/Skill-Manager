package usecase

import (
	"context"
	"fmt"
	"os"
)

// ResetProjectSkills removes all skill directories copied into a project's agent
// skill folders (.claude/skills/ and .github/skills/) and deletes all activations
// scoped to that project.
type ResetProjectSkills struct {
	projects    ProjectRepository
	activations ActivationRepository
}

func NewResetProjectSkills(projects ProjectRepository, activations ActivationRepository) *ResetProjectSkills {
	return &ResetProjectSkills{projects: projects, activations: activations}
}

func (uc *ResetProjectSkills) Execute(ctx context.Context, projectID string) error {
	project, err := uc.projects.GetByID(ctx, projectID)
	if err != nil {
		return fmt.Errorf("reset project skills: get project: %w", err)
	}

	for _, agent := range []string{"claude", "copilot"} {
		dir := agentSkillsDir(project.Path, agent)
		if err := os.RemoveAll(dir); err != nil {
			return fmt.Errorf("reset project skills: remove %s: %w", dir, err)
		}
	}

	if err := uc.activations.DeleteByProjectID(ctx, projectID); err != nil {
		return fmt.Errorf("reset project skills: clear activations: %w", err)
	}

	return nil
}
