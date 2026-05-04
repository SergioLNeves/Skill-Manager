package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"skill-manager/internal/binding"
	"skill-manager/internal/config"
	"skill-manager/internal/di"
	"skill-manager/internal/logger"
	"skill-manager/internal/watcher"
)

// App is the Wails application struct.
// All public methods are exposed to the frontend via the generated wailsjs bindings.
type App struct {
	ctx         context.Context
	cancel      context.CancelFunc
	container   *di.Container
	logCleanup  func()
	settingsPath string
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx, a.cancel = context.WithCancel(ctx)

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "skills-manager: user home dir: %v\n", err)
		return
	}

	base := filepath.Join(home, ".skills-manager")
	a.settingsPath = filepath.Join(base, "settings.json")

	// Structured logging.
	cleanup, err := logger.Init(filepath.Join(base, "logs"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "skills-manager: init logger: %v\n", err)
	}
	a.logCleanup = cleanup
	slog.Info("skills-manager starting")

	// Settings.
	cfg, err := config.Load(a.settingsPath, config.DefaultSettings(home))
	if err != nil {
		slog.Warn("failed to load settings, using defaults", "err", err)
		cfg = config.DefaultSettings(home)
	}

	skillsHome := cfg.SkillsHome
	if env := os.Getenv("SKILLS_MANAGER_HOME"); env != "" {
		skillsHome = env
		slog.Info("overriding skills home from env", "path", skillsHome)
	}

	if err = os.MkdirAll(skillsHome, 0o755); err != nil {
		slog.Error("cannot create skills home", "path", skillsHome, "err", err)
		return
	}

	dbPath := filepath.Join(base, "registry.db")
	container, err := di.Wire(skillsHome, dbPath)
	if err != nil {
		slog.Error("wire container failed", "err", err)
		return
	}
	a.container = container

	// File watcher for skills directory.
	sw, err := watcher.NewSkillsWatcher(skillsHome, func() {
		slog.Info("skills directory changed")
		// The frontend re-fetches via TanStack Query on the next focus/poll;
		// future work: push a Wails event here.
	})
	if err != nil {
		slog.Warn("skills watcher unavailable", "err", err)
	} else {
		go sw.Run(a.ctx)
	}

	slog.Info("skills-manager ready", "skillsHome", skillsHome, "db", dbPath)
}

func (a *App) shutdown(_ context.Context) {
	if a.cancel != nil {
		a.cancel()
	}
	if a.container != nil && a.container.DB != nil {
		a.container.DB.Close()
	}
	if a.logCleanup != nil {
		a.logCleanup()
	}
}

// --- Settings ---

// GetSettings returns the current user settings.
func (a *App) GetSettings() (config.Settings, error) {
	home, _ := os.UserHomeDir()
	return config.Load(a.settingsPath, config.DefaultSettings(home))
}

// SaveSettings persists updated settings.
func (a *App) SaveSettings(s config.Settings) error {
	if err := config.Save(a.settingsPath, s); err != nil {
		return fmt.Errorf("save settings: %w", err)
	}
	slog.Info("settings saved", "workspaceRoots", s.WorkspaceRoots)
	return nil
}

func (a *App) ready() error {
	if a.container == nil {
		return fmt.Errorf("app not initialized: check logs for startup errors")
	}
	return nil
}

// --- Skills ---

func (a *App) ListSkills() ([]binding.SkillDTO, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	return a.container.Skills.List(a.ctx)
}

// --- Projects ---

func (a *App) ListProjects() ([]binding.ProjectDTO, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	return a.container.Projects.List(a.ctx)
}

func (a *App) RegisterProject(req binding.RegisterProjectRequestDTO) (binding.ProjectDTO, error) {
	if err := a.ready(); err != nil {
		return binding.ProjectDTO{}, err
	}
	return a.container.Projects.RegisterManual(a.ctx, req)
}

func (a *App) ScanCandidates(roots []string) ([]binding.ProjectCandidateDTO, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	return a.container.Projects.ScanCandidates(a.ctx, roots)
}

func (a *App) ConfirmCandidate(candidate binding.ProjectCandidateDTO) (binding.ProjectDTO, error) {
	if err := a.ready(); err != nil {
		return binding.ProjectDTO{}, err
	}
	return a.container.Projects.ConfirmCandidate(a.ctx, candidate)
}

func (a *App) DeleteProject(id string) error {
	if err := a.ready(); err != nil {
		return err
	}
	return a.container.Projects.Delete(a.ctx, id)
}

// --- Activations ---

func (a *App) ListActivations(filter binding.ActivationFilterDTO) ([]binding.ActivationDTO, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	return a.container.Activations.List(a.ctx, filter)
}

func (a *App) Activate(req binding.ActivateRequestDTO) (binding.ActivateResultDTO, error) {
	if err := a.ready(); err != nil {
		return binding.ActivateResultDTO{}, err
	}
	return a.container.Activations.Activate(a.ctx, req)
}

func (a *App) Deactivate(id int64) error {
	if err := a.ready(); err != nil {
		return err
	}
	return a.container.Activations.Deactivate(a.ctx, id)
}

func (a *App) ResolveConflict(req binding.ResolveConflictRequestDTO) error {
	if err := a.ready(); err != nil {
		return err
	}
	return a.container.Activations.ResolveConflict(a.ctx, req)
}

// --- Doctor ---

func (a *App) RunDoctor() (binding.DoctorReportDTO, error) {
	if err := a.ready(); err != nil {
		return binding.DoctorReportDTO{}, err
	}
	return a.container.Doctor.Run(a.ctx)
}
