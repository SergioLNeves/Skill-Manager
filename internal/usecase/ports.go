package usecase

import (
	"context"

	"skill-manager/internal/domain"
)

// SkillRepository reads global skills from configured source directories.
type SkillRepository interface {
	List(ctx context.Context) ([]domain.Skill, error)
	GetByID(ctx context.Context, id string) (domain.Skill, error)
}

// ProjectSkillRepository discovers skills that live inside a project directory.
type ProjectSkillRepository interface {
	ListByProject(ctx context.Context, project domain.Project) ([]domain.Skill, error)
	GetByID(ctx context.Context, id string, project domain.Project) (domain.Skill, error)
}

// ProjectRepository persists registered projects.
type ProjectRepository interface {
	List(ctx context.Context) ([]domain.Project, error)
	GetByID(ctx context.Context, id string) (domain.Project, error)
	Save(ctx context.Context, p domain.Project) error
	Delete(ctx context.Context, id string) error
}

// ActivationFilter narrows the activation query.
type ActivationFilter struct {
	SkillID   string
	Agent     domain.Agent
	Scope     domain.Scope
	ProjectID string
}

// ActivationRepository persists skill activations.
type ActivationRepository interface {
	List(ctx context.Context, filter ActivationFilter) ([]domain.Activation, error)
	Save(ctx context.Context, a domain.Activation) (domain.Activation, error)
	Delete(ctx context.Context, id int64) error
	FindConflict(ctx context.Context, skillID string, agent domain.Agent, projectID string) (*domain.Conflict, error)
}

// AgentAdapter applies the desired activation state to the agent's filesystem paths.
type AgentAdapter interface {
	Agent() domain.Agent
	ApplyGlobal(ctx context.Context, activeSkills []domain.Skill) error
	ApplyProject(ctx context.Context, project domain.Project, activeSkills []domain.Skill) error
	CapabilityCheck(ctx context.Context) error
}

// ProjectCandidate is a project detected by the scanner but not yet registered.
type ProjectCandidate struct {
	Name           string
	Path           string
	DetectedAgents []domain.Agent
}

// ProjectScanner detects project directories within given workspace roots.
type ProjectScanner interface {
	Scan(ctx context.Context, roots []string) ([]ProjectCandidate, error)
}

// DoctorIssue describes a detected inconsistency.
type DoctorIssue struct {
	Kind     string            // machine-readable category
	Title    string            // short human-readable title
	Detail   string            // full description of what is wrong
	HowToFix string            // what the auto-fix will do
	Fixable  bool
	FixData  map[string]string // data needed to execute the fix
}

// DoctorReport is the result of a health check run.
type DoctorReport struct {
	Issues []DoctorIssue
}
