package binding

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"skill-manager/internal/domain"
	"skill-manager/internal/hashid"
	"skill-manager/internal/usecase"
)

// ProjectsBinding exposes project-related operations to the Wails frontend.
type ProjectsBinding struct {
	listProjects    *usecase.ListProjects
	registerProject *usecase.RegisterProject
	scanProjects    *usecase.ScanProjects
	deleteProject   *usecase.DeleteProject
}

func NewProjectsBinding(
	list *usecase.ListProjects,
	register *usecase.RegisterProject,
	scan *usecase.ScanProjects,
	del *usecase.DeleteProject,
) *ProjectsBinding {
	return &ProjectsBinding{
		listProjects:    list,
		registerProject: register,
		scanProjects:    scan,
		deleteProject:   del,
	}
}

// List returns all registered projects.
func (b *ProjectsBinding) List(ctx context.Context) ([]ProjectDTO, error) {
	projects, err := b.listProjects.Execute(ctx)
	if err != nil {
		return nil, fmt.Errorf("projects binding: list: %w", err)
	}
	dtos := make([]ProjectDTO, len(projects))
	for i, p := range projects {
		dtos[i] = toProjectDTO(p)
	}
	return dtos, nil
}

// RegisterManual registers a project by its filesystem path.
func (b *ProjectsBinding) RegisterManual(ctx context.Context, req RegisterProjectRequestDTO) (ProjectDTO, error) {
	agents := make([]domain.Agent, len(req.DetectedAgents))
	for i, a := range req.DetectedAgents {
		agents[i] = domain.Agent(a)
	}
	p := domain.Project{
		ID:             hashid.Stable(req.Path),
		Name:           filepath.Base(req.Path),
		Path:           req.Path,
		DetectedAgents: agents,
		AddedAt:        time.Now(),
	}
	if err := b.registerProject.Execute(ctx, p); err != nil {
		return ProjectDTO{}, fmt.Errorf("projects binding: register: %w", err)
	}
	return toProjectDTO(p), nil
}

// ScanCandidates scans workspace roots and returns detected project candidates.
func (b *ProjectsBinding) ScanCandidates(ctx context.Context, roots []string) ([]ProjectCandidateDTO, error) {
	candidates, err := b.scanProjects.Execute(ctx, roots)
	if err != nil {
		return nil, fmt.Errorf("projects binding: scan: %w", err)
	}
	dtos := make([]ProjectCandidateDTO, len(candidates))
	for i, c := range candidates {
		dtos[i] = toCandidateDTO(c)
	}
	return dtos, nil
}

// ConfirmCandidate registers a scanned candidate as a project.
func (b *ProjectsBinding) ConfirmCandidate(ctx context.Context, candidate ProjectCandidateDTO) (ProjectDTO, error) {
	return b.RegisterManual(ctx, RegisterProjectRequestDTO{
		Path:           candidate.Path,
		DetectedAgents: candidate.DetectedAgents,
	})
}

// Delete removes a registered project.
func (b *ProjectsBinding) Delete(ctx context.Context, id string) error {
	if err := b.deleteProject.Execute(ctx, id); err != nil {
		return fmt.Errorf("projects binding: delete %s: %w", id, err)
	}
	return nil
}
