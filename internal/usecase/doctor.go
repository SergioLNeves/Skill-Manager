package usecase

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"skill-manager/internal/domain"
)

// Doctor validates the consistency between the DB state and the filesystem.
type Doctor struct {
	skills      SkillRepository
	projects    ProjectRepository
	activations ActivationRepository
	homeDir     string // user home, used to locate managed symlink directories
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
					Detail:  fmt.Sprintf("skill %q no longer exists (activation id %d)", a.SkillID, a.ID),
					Fixable: true,
				})
				continue
			}
			return fmt.Errorf("check skill %s: %w", a.SkillID, err)
		}
	}
	return nil
}

// checkInvalidProjects finds project IDs referenced by activations that no longer exist in the DB.
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
					Detail:  fmt.Sprintf("project %q referenced by activations no longer exists", id),
					Fixable: true,
				})
				continue
			}
			return fmt.Errorf("check project %s: %w", id, err)
		}
	}
	return nil
}

// checkBrokenSymlinks looks for managed symlinks in ~/.claude/skills/ and
// project .claude/skills/ directories that point to a non-existent source.
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
					Kind:    "broken_symlink",
					Detail:  fmt.Sprintf("symlink %s points to missing path %s", linkPath, dest),
					Fixable: true,
				})
			}
		}
	}
	return nil
}

// checkProjectPaths verifies that each registered project's path still exists on disk.
func (uc *Doctor) checkProjectPaths(ctx context.Context, report *DoctorReport) error {
	projects, err := uc.projects.List(ctx)
	if err != nil {
		return fmt.Errorf("list projects: %w", err)
	}

	for _, p := range projects {
		if _, err := os.Stat(p.Path); os.IsNotExist(err) {
			report.Issues = append(report.Issues, DoctorIssue{
				Kind:    "missing_project_path",
				Detail:  fmt.Sprintf("project %q path %q no longer exists on disk", p.Name, p.Path),
				Fixable: false,
			})
		}
	}
	return nil
}

// hasCopilotMarkers checks whether a file contains both managed block markers.
func hasCopilotMarkers(content string) bool {
	return strings.Contains(content, "<!-- skills-manager:start -->") &&
		strings.Contains(content, "<!-- skills-manager:end -->")
}
