package usecase

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"skill-manager/internal/domain"
)

// Doctor validates the consistency between the DB state and the filesystem.
type Doctor struct {
	skills      SkillRepository
	projects    ProjectRepository
	activations ActivationRepository
	homeDir     string
}

func NewDoctor(
	skills SkillRepository,
	projects ProjectRepository,
	activations ActivationRepository,
	homeDir string,
) *Doctor {
	return &Doctor{
		skills:      skills,
		projects:    projects,
		activations: activations,
		homeDir:     homeDir,
	}
}

func (uc *Doctor) Execute(ctx context.Context) (DoctorReport, error) {
	var report DoctorReport

	if err := uc.checkOrphanedActivations(ctx, &report); err != nil {
		return DoctorReport{}, fmt.Errorf("doctor: %w", err)
	}
	if err := uc.checkInvalidProjects(ctx, &report); err != nil {
		return DoctorReport{}, fmt.Errorf("doctor: %w", err)
	}
	if err := uc.checkBrokenSymlinks(ctx, &report); err != nil {
		return DoctorReport{}, fmt.Errorf("doctor: %w", err)
	}
	if err := uc.checkProjectPaths(ctx, &report); err != nil {
		return DoctorReport{}, fmt.Errorf("doctor: %w", err)
	}

	return report, nil
}

func (uc *Doctor) checkOrphanedActivations(ctx context.Context, report *DoctorReport) error {
	all, err := uc.activations.List(ctx, ActivationFilter{})
	if err != nil {
		return fmt.Errorf("list activations: %w", err)
	}

	for _, a := range all {
		if _, err := uc.skills.GetByID(ctx, a.SkillID); err != nil {
			if !errors.Is(err, domain.ErrSkillNotFound) {
				return fmt.Errorf("check skill %s: %w", a.SkillID, err)
			}
			agentStr := string(a.Agent)
			report.Issues = append(report.Issues, DoctorIssue{
				Kind:     "orphaned_activation",
				Title:    "Ativação órfã",
				Detail:   fmt.Sprintf("A ativação para o agente %q (id %d) referencia uma skill que não existe mais no disco.", agentStr, a.ID),
				HowToFix: "Remove o registro de ativação do banco de dados.",
				Fixable:  true,
				FixData:  map[string]string{"activation_id": fmt.Sprintf("%d", a.ID)},
			})
		}
	}
	return nil
}

func (uc *Doctor) checkInvalidProjects(ctx context.Context, report *DoctorReport) error {
	activations, err := uc.activations.List(ctx, ActivationFilter{Scope: domain.ScopeProject})
	if err != nil {
		return fmt.Errorf("list project activations: %w", err)
	}

	checked := map[string]int{}
	for _, a := range activations {
		if a.ProjectID == nil {
			continue
		}
		checked[*a.ProjectID]++
	}

	for projectID, count := range checked {
		if _, err := uc.projects.GetByID(ctx, projectID); err != nil {
			if !errors.Is(err, domain.ErrProjectNotFound) {
				return fmt.Errorf("check project %s: %w", projectID, err)
			}
			report.Issues = append(report.Issues, DoctorIssue{
				Kind:     "missing_project",
				Title:    "Projeto removido com ativações pendentes",
				Detail:   fmt.Sprintf("O projeto foi removido do registro, mas ainda existem %d ativação(ões) referenciando-o.", count),
				HowToFix: "Remove as ativações órfãs do banco de dados.",
				Fixable:  true,
				FixData:  map[string]string{"project_id": projectID},
			})
		}
	}
	return nil
}

func (uc *Doctor) checkBrokenSymlinks(ctx context.Context, report *DoctorReport) error {
	dirs := []string{filepath.Join(uc.homeDir, ".claude", "skills")}

	projects, err := uc.projects.List(ctx)
	if err != nil {
		return fmt.Errorf("list projects: %w", err)
	}
	for _, p := range projects {
		dirs = append(dirs, filepath.Join(p.Path, ".claude", "skills"))
	}

	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return fmt.Errorf("read dir %s: %w", dir, err)
		}

		for _, e := range entries {
			if e.Type()&os.ModeSymlink == 0 {
				continue
			}
			linkPath := filepath.Join(dir, e.Name())
			dest, err := os.Readlink(linkPath)
			if err != nil {
				continue
			}
			if _, statErr := os.Stat(dest); os.IsNotExist(statErr) {
				report.Issues = append(report.Issues, DoctorIssue{
					Kind:     "broken_symlink",
					Title:    "Symlink quebrado",
					Detail:   fmt.Sprintf("O symlink %q aponta para %q que não existe mais no disco.", linkPath, dest),
					HowToFix: "Remove o symlink quebrado. A skill precisará ser reativada se quiser usá-la novamente.",
					Fixable:  true,
					FixData:  map[string]string{"symlink_path": linkPath},
				})
			}
		}
	}
	return nil
}

func (uc *Doctor) checkProjectPaths(ctx context.Context, report *DoctorReport) error {
	projects, err := uc.projects.List(ctx)
	if err != nil {
		return fmt.Errorf("list projects: %w", err)
	}

	for _, p := range projects {
		if _, err := os.Stat(p.Path); os.IsNotExist(err) {
			report.Issues = append(report.Issues, DoctorIssue{
				Kind:     "missing_project_path",
				Title:    "Diretório do projeto não encontrado",
				Detail:   fmt.Sprintf("O projeto %q está registrado, mas o diretório %q não existe mais no disco.", p.Name, p.Path),
				HowToFix: "Remove o projeto do registro. Você pode adicioná-lo novamente depois se o diretório for restaurado.",
				Fixable:  true,
				FixData:  map[string]string{"project_id": p.ID, "project_name": p.Name},
			})
		}
	}
	return nil
}
