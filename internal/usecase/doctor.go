package usecase

import (
	"context"
	"errors"
	"fmt"

	"skill-manager/internal/domain"
)

// Doctor validates the consistency between the DB state and the filesystem.
type Doctor struct {
	skills      SkillRepository
	projects    ProjectRepository
	activations ActivationRepository
}

func NewDoctor(skills SkillRepository, projects ProjectRepository, activations ActivationRepository) *Doctor {
	return &Doctor{
		skills:      skills,
		projects:    projects,
		activations: activations,
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

	return report, nil
}

// checkOrphanedActivations finds activations whose skill no longer exists.
func (uc *Doctor) checkOrphanedActivations(ctx context.Context, report *DoctorReport) error {
	all, err := uc.activations.List(ctx, ActivationFilter{})
	if err != nil {
		return fmt.Errorf("list activations: %w", err)
	}

	for _, a := range all {
		if _, err := uc.skills.GetByID(ctx, a.SkillID); err != nil {
			if errors.Is(err, domain.ErrSkillNotFound) {
				report.Issues = append(report.Issues, DoctorIssue{
					Kind:    "orphaned_activation",
					Detail:  fmt.Sprintf("skill %s no longer exists (activation id %d)", a.SkillID, a.ID),
					Fixable: true,
				})
				continue
			}
			return fmt.Errorf("check skill %s: %w", a.SkillID, err)
		}
	}
	return nil
}

// checkInvalidProjects finds projects whose path no longer exists in the DB.
func (uc *Doctor) checkInvalidProjects(ctx context.Context, report *DoctorReport) error {
	activations, err := uc.activations.List(ctx, ActivationFilter{Scope: domain.ScopeProject})
	if err != nil {
		return fmt.Errorf("list project activations: %w", err)
	}

	checked := map[string]bool{}
	for _, a := range activations {
		if a.ProjectID == nil {
			continue
		}
		id := *a.ProjectID
		if checked[id] {
			continue
		}
		checked[id] = true

		if _, err := uc.projects.GetByID(ctx, id); err != nil {
			if errors.Is(err, domain.ErrProjectNotFound) {
				report.Issues = append(report.Issues, DoctorIssue{
					Kind:    "missing_project",
					Detail:  fmt.Sprintf("project %s referenced by activations no longer exists", id),
					Fixable: true,
				})
				continue
			}
			return fmt.Errorf("check project %s: %w", id, err)
		}
	}
	return nil
}
