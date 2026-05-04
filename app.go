package main

import (
	"context"
	"fmt"
	"os"

	"skill-manager/internal/binding"
	"skill-manager/internal/di"
)

// App is the Wails application struct. Its public methods are exposed to the frontend.
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

	if err := os.MkdirAll(skillsHome, 0o755); err != nil {
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

// Skills returns the skills binding for Wails method exposure.
func (a *App) Skills() *binding.SkillsBinding {
	return a.container.Skills
}

// Projects returns the projects binding.
func (a *App) Projects() *binding.ProjectsBinding {
	return a.container.Projects
}

// Activations returns the activation binding.
func (a *App) Activations() *binding.ActivationBinding {
	return a.container.Activations
}

// Doctor returns the doctor binding.
func (a *App) Doctor() *binding.DoctorBinding {
	return a.container.Doctor
}
