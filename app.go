package main

import (
	"context"
	"fmt"
	"os"

	"skill-manager/internal/binding"
	"skill-manager/internal/di"
)

// App is the Wails application struct.
// All public methods are exposed to the frontend via the generated wailsjs bindings.
type App struct {
	ctx       context.Context
	container *di.Container
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	skillsHome, dbPath, err := di.DefaultPaths()
	if err != nil {
		fmt.Fprintf(os.Stderr, "skills-manager: resolve paths: %v\n", err)
		return
	}
	if err = os.MkdirAll(skillsHome, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "skills-manager: mkdir skills home: %v\n", err)
		return
	}

	container, err := di.Wire(skillsHome, dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "skills-manager: wire container: %v\n", err)
		return
	}
	a.container = container
}

func (a *App) shutdown(_ context.Context) {
	if a.container != nil && a.container.DB != nil {
		a.container.DB.Close()
	}
}

// --- Skills ---

func (a *App) ListSkills() ([]binding.SkillDTO, error) {
	return a.container.Skills.List(a.ctx)
}

// --- Projects ---

func (a *App) ListProjects() ([]binding.ProjectDTO, error) {
	return a.container.Projects.List(a.ctx)
}

func (a *App) RegisterProject(req binding.RegisterProjectRequestDTO) (binding.ProjectDTO, error) {
	return a.container.Projects.RegisterManual(a.ctx, req)
}

func (a *App) ScanCandidates(roots []string) ([]binding.ProjectCandidateDTO, error) {
	return a.container.Projects.ScanCandidates(a.ctx, roots)
}

func (a *App) ConfirmCandidate(candidate binding.ProjectCandidateDTO) (binding.ProjectDTO, error) {
	return a.container.Projects.ConfirmCandidate(a.ctx, candidate)
}

func (a *App) DeleteProject(id string) error {
	return a.container.Projects.Delete(a.ctx, id)
}

// --- Activations ---

func (a *App) ListActivations(filter binding.ActivationFilterDTO) ([]binding.ActivationDTO, error) {
	return a.container.Activations.List(a.ctx, filter)
}

func (a *App) Activate(req binding.ActivateRequestDTO) (binding.ActivateResultDTO, error) {
	return a.container.Activations.Activate(a.ctx, req)
}

func (a *App) Deactivate(id int64) error {
	return a.container.Activations.Deactivate(a.ctx, id)
}

func (a *App) ResolveConflict(req binding.ResolveConflictRequestDTO) error {
	return a.container.Activations.ResolveConflict(a.ctx, req)
}

// --- Doctor ---

func (a *App) RunDoctor() (binding.DoctorReportDTO, error) {
	return a.container.Doctor.Run(a.ctx)
}
