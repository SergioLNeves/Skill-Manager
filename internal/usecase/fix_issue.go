package usecase

import (
	"context"
	"fmt"
	"os"
	"strconv"
)

// FixIssue applies the automated fix for a single DoctorIssue.
type FixIssue struct {
	activations ActivationRepository
	projects    ProjectRepository
}

func NewFixIssue(activations ActivationRepository, projects ProjectRepository) *FixIssue {
	return &FixIssue{activations: activations, projects: projects}
}

func (uc *FixIssue) Execute(ctx context.Context, issue DoctorIssue) error {
	if !issue.Fixable {
		return fmt.Errorf("fix issue: issue %q is not fixable", issue.Kind)
	}

	switch issue.Kind {
	case "orphaned_activation":
		return uc.deleteActivation(ctx, issue.FixData)
	case "missing_project":
		return uc.deleteProjectActivations(ctx, issue.FixData)
	case "broken_symlink":
		return uc.removeBrokenSymlink(issue.FixData)
	case "missing_project_path":
		return uc.deleteProject(ctx, issue.FixData)
	default:
		return fmt.Errorf("fix issue: unknown kind %q", issue.Kind)
	}
}

func (uc *FixIssue) deleteActivation(ctx context.Context, data map[string]string) error {
	id, err := strconv.ParseInt(data["activation_id"], 10, 64)
	if err != nil {
		return fmt.Errorf("fix orphaned_activation: parse id: %w", err)
	}
	if err := uc.activations.Delete(ctx, id); err != nil {
		return fmt.Errorf("fix orphaned_activation: %w", err)
	}
	return nil
}

func (uc *FixIssue) deleteProjectActivations(ctx context.Context, data map[string]string) error {
	projectID := data["project_id"]
	all, err := uc.activations.List(ctx, ActivationFilter{ProjectID: projectID})
	if err != nil {
		return fmt.Errorf("fix missing_project: list activations: %w", err)
	}
	for _, a := range all {
		_ = uc.activations.Delete(ctx, a.ID)
	}
	return nil
}

func (uc *FixIssue) removeBrokenSymlink(data map[string]string) error {
	path := data["symlink_path"]
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("fix broken_symlink: remove %s: %w", path, err)
	}
	return nil
}

func (uc *FixIssue) deleteProject(ctx context.Context, data map[string]string) error {
	if err := uc.projects.Delete(ctx, data["project_id"]); err != nil {
		return fmt.Errorf("fix missing_project_path: %w", err)
	}
	return nil
}
