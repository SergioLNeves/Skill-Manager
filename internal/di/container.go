package di

import (
	"database/sql"
	"os"
	"path/filepath"

	"github.com/spf13/afero"

	"skill-manager/internal/adapter/agent"
	"skill-manager/internal/adapter/filesystem"
	"skill-manager/internal/adapter/persistence"
	"skill-manager/internal/binding"
	"skill-manager/internal/domain"
	"skill-manager/internal/usecase"
)

// Container holds all wired dependencies for the application.
type Container struct {
	DB *sql.DB

	Skills      *binding.SkillsBinding
	Projects    *binding.ProjectsBinding
	Activations *binding.ActivationBinding
	Doctor      *binding.DoctorBinding
}

// Wire builds the full dependency graph and returns a ready Container.
// skillsHome is the path to the central skills repository (e.g. ~/.skills-manager/skills).
// dbPath is the path to the SQLite registry (e.g. ~/.skills-manager/registry.db).
func Wire(skillsHome, dbPath string) (*Container, error) {
	db, err := persistence.Open(dbPath)
	if err != nil {
		return nil, err
	}

	// Repositories
	skillRepo := filesystem.NewSkillRepository(afero.NewOsFs(), skillsHome)
	projectRepo := persistence.NewProjectRepository(db)
	activationRepo := persistence.NewActivationRepository(db)
	projectScanner := filesystem.NewProjectScanner(afero.NewOsFs())

	// Agent adapters
	homeDir, err := os.UserHomeDir()
	if err != nil {
		db.Close()
		return nil, err
	}
	adapters := map[domain.Agent]usecase.AgentAdapter{
		domain.AgentClaude:  agent.NewClaudeAdapter(homeDir),
		domain.AgentCopilot: agent.NewCopilotAdapter(),
	}

	// Use cases
	listSkills := usecase.NewListSkills(skillRepo)
	listProjects := usecase.NewListProjects(projectRepo)
	registerProject := usecase.NewRegisterProject(projectRepo)
	scanProjects := usecase.NewScanProjects(projectScanner)
	deleteProject := usecase.NewDeleteProject(projectRepo)
	activateSkill := usecase.NewActivateSkill(skillRepo, projectRepo, activationRepo, adapters)
	deactivateSkill := usecase.NewDeactivateSkill(skillRepo, projectRepo, activationRepo, adapters)
	resolveConflict := usecase.NewResolveConflict(projectRepo, skillRepo, activationRepo, adapters)
	doctor := usecase.NewDoctor(skillRepo, projectRepo, activationRepo)

	return &Container{
		DB:          db,
		Skills:      binding.NewSkillsBinding(listSkills),
		Projects:    binding.NewProjectsBinding(listProjects, registerProject, scanProjects, deleteProject),
		Activations: binding.NewActivationBinding(activateSkill, deactivateSkill, resolveConflict, activationRepo),
		Doctor:      binding.NewDoctorBinding(doctor),
	}, nil
}

// DefaultPaths returns the default skillsHome and dbPath under the user home.
func DefaultPaths() (skillsHome, dbPath string, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", err
	}
	base := filepath.Join(home, ".skills-manager")
	return filepath.Join(base, "skills"), filepath.Join(base, "registry.db"), nil
}
