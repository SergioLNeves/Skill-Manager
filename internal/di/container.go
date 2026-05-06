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
	Categories  *binding.CategoriesBinding

	RefreshCache *usecase.RefreshSkillCache
}

// Wire builds the full dependency graph and returns a ready Container.
func Wire(globalSkillSources []string, dbPath string) (*Container, error) {
	db, err := persistence.Open(dbPath)
	if err != nil {
		return nil, err
	}

	// Repositories
	skillRepo := filesystem.NewSkillRepository(afero.NewOsFs(), globalSkillSources...)
	projectSkillRepo := filesystem.NewProjectSkillRepository(afero.NewOsFs())
	projectRepo := persistence.NewProjectRepository(db)
	activationRepo := persistence.NewActivationRepository(db)
	projectScanner := filesystem.NewProjectScanner(afero.NewOsFs())
	skillCacheRepo := persistence.NewSkillCacheRepository(db, projectRepo)
	categoryRepo := persistence.NewCategoryRepository(db)

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
	listProjectSkills := usecase.NewListProjectSkills(projectRepo, projectSkillRepo)
	refreshCache := usecase.NewRefreshSkillCache(skillRepo, projectRepo, projectSkillRepo, skillCacheRepo)
	listAllSkills := usecase.NewListAllSkills(refreshCache, skillCacheRepo)
	copySkill := usecase.NewCopySkill(skillRepo, projectRepo, projectSkillRepo)
	deleteSkill := usecase.NewDeleteSkill(skillRepo, projectRepo, projectSkillRepo, activationRepo)
	listProjects := usecase.NewListProjects(projectRepo)
	registerProject := usecase.NewRegisterProject(projectRepo)
	scanProjects := usecase.NewScanProjects(projectScanner)
	deleteProject := usecase.NewDeleteProject(projectRepo)
	activateSkill := usecase.NewActivateSkill(skillRepo, projectSkillRepo, projectRepo, activationRepo, adapters)
	deactivateSkill := usecase.NewDeactivateSkill(skillRepo, projectSkillRepo, projectRepo, activationRepo, adapters)
	resolveConflict := usecase.NewResolveConflict(projectRepo, skillRepo, projectSkillRepo, activationRepo, adapters)
	resetProjectSkills := usecase.NewResetProjectSkills(projectRepo, activationRepo)
	doctor := usecase.NewDoctor(skillRepo, projectRepo, activationRepo, homeDir)
	fixIssue := usecase.NewFixIssue(activationRepo, projectRepo)
	manageCategories := usecase.NewManageCategories(categoryRepo)
	assignSkillCategory := usecase.NewAssignSkillCategory(categoryRepo, projectRepo, activationRepo, skillRepo, projectSkillRepo, adapters)
	associateProjectCategory := usecase.NewAssociateProjectCategory(categoryRepo, projectRepo, activationRepo, skillRepo, projectSkillRepo, adapters)
	disassociateProjectCategory := usecase.NewDisassociateProjectCategory(categoryRepo)

	return &Container{
		DB:           db,
		RefreshCache: refreshCache,
		Skills:       binding.NewSkillsBinding(listSkills, listProjectSkills, listAllSkills, copySkill, deleteSkill, resetProjectSkills),
		Projects:     binding.NewProjectsBinding(listProjects, registerProject, scanProjects, deleteProject, activationRepo),
		Activations:  binding.NewActivationBinding(activateSkill, deactivateSkill, resolveConflict, activationRepo),
		Doctor:       binding.NewDoctorBinding(doctor, fixIssue),
		Categories:   binding.NewCategoriesBinding(manageCategories, assignSkillCategory, associateProjectCategory, disassociateProjectCategory, categoryRepo),
	}, nil
}

// DefaultDBPath returns the default database path under the user home.
func DefaultDBPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".skills-manager", "registry.db"), nil
}
