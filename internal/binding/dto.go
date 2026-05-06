package binding

import (
	"time"

	"skill-manager/internal/domain"
	"skill-manager/internal/usecase"
)

// --- Category DTOs ---

// CategoryDTO is the frontend representation of a category.
type CategoryDTO struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"createdAt"`
}

// CreateCategoryRequestDTO carries a category creation request from the frontend.
type CreateCategoryRequestDTO struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UpdateCategoryRequestDTO carries a category update request from the frontend.
type UpdateCategoryRequestDTO struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// AssignSkillCategoryRequestDTO sets or clears the category of a skill.
type AssignSkillCategoryRequestDTO struct {
	SkillName  string  `json:"skillName"`
	SkillPath  string  `json:"skillPath"`
	CategoryID *int64  `json:"categoryId"`
}

// ProjectCategoryRequestDTO carries a project-category-agent association request.
type ProjectCategoryRequestDTO struct {
	ProjectID  string `json:"projectId"`
	CategoryID int64  `json:"categoryId"`
	Agent      string `json:"agent"`
}

// ProjectCategoryLinkDTO is the frontend representation of a project-category-agent link.
type ProjectCategoryLinkDTO struct {
	ProjectID  string      `json:"projectId"`
	CategoryID int64       `json:"categoryId"`
	Agent      string      `json:"agent"`
	Category   CategoryDTO `json:"category"`
}

// SkillDTO is the frontend representation of a skill.
type SkillDTO struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	CategoryID       *int64 `json:"categoryId"`
	CategoryName     string `json:"categoryName"`
	Path             string `json:"path"`
	Source           string `json:"source"`           // "global" or "project"
	OwnerProjectID   string `json:"ownerProjectId"`   // non-empty when Source == "project"
	OwnerProjectName string `json:"ownerProjectName"` // display name of owning project
	UpdatedAt        string `json:"updatedAt"`
}

// AggregatedSkillDTO is the frontend representation of a deduplicated skill with all locations.
type AggregatedSkillDTO struct {
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	CategoryID   *int64            `json:"categoryId"`
	CategoryName string            `json:"categoryName"`
	IsGlobal     bool              `json:"isGlobal"`
	GlobalPath   string            `json:"globalPath"`
	Projects     []SkillProjectRef `json:"projects"`
	UpdatedAt    string            `json:"updatedAt"`
}

// SkillProjectRef is a lightweight project reference inside an aggregated skill.
type SkillProjectRef struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Path      string `json:"path"`      // project root path
	SkillPath string `json:"skillPath"` // skill directory path (for reading SKILL.md)
}

// CopySkillRequestDTO carries a skill copy request from the frontend.
type CopySkillRequestDTO struct {
	SkillID         string `json:"skillId"`
	SourceProjectID string `json:"sourceProjectId"`
	TargetProjectID string `json:"targetProjectId"`
	Agent           string `json:"agent"` // "claude" or "copilot"
}

// DeleteSkillRequestDTO carries a skill deletion request from the frontend.
type DeleteSkillRequestDTO struct {
	SkillID   string `json:"skillId"`
	ProjectID string `json:"projectId"`
}

// ProjectDTO is the frontend representation of a registered project.
type ProjectDTO struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Path           string   `json:"path"`
	DetectedAgents []string `json:"detectedAgents"`
	ActiveAgents   []string `json:"activeAgents"` // agents with at least one active skill
	AddedAt        string   `json:"addedAt"`
}

// ProjectCandidateDTO is a project detected by the scanner but not yet registered.
type ProjectCandidateDTO struct {
	Name           string   `json:"name"`
	Path           string   `json:"path"`
	DetectedAgents []string `json:"detectedAgents"`
}

// ActivationDTO is the frontend representation of an activation record.
type ActivationDTO struct {
	ID        int64  `json:"id"`
	SkillID   string `json:"skillId"`
	Agent     string `json:"agent"`
	Scope     string `json:"scope"`
	ProjectID string `json:"projectId"` // empty string when global
	AppliedAt string `json:"appliedAt"`
}

// ActivationFilterDTO narrows list queries from the frontend.
type ActivationFilterDTO struct {
	SkillID   string `json:"skillId"`
	Agent     string `json:"agent"`
	Scope     string `json:"scope"`
	ProjectID string `json:"projectId"`
}

// ActivateRequestDTO carries a skill activation request from the frontend.
type ActivateRequestDTO struct {
	SkillID   string `json:"skillId"`
	Agent     string `json:"agent"`
	Scope     string `json:"scope"`
	ProjectID string `json:"projectId"`
}

// ConflictDTO describes a detected conflict between global and project activations.
type ConflictDTO struct {
	SkillID           string         `json:"skillId"`
	Agent             string         `json:"agent"`
	GlobalActivation  *ActivationDTO `json:"globalActivation"`
	ProjectActivation *ActivationDTO `json:"projectActivation"`
}

// ActivateResultDTO is the response from an Activate call.
// Exactly one of Activation or Conflict will be non-nil.
type ActivateResultDTO struct {
	Activation *ActivationDTO `json:"activation"`
	Conflict   *ConflictDTO   `json:"conflict"`
}

// ResolveConflictRequestDTO carries the resolution choice from the frontend.
type ResolveConflictRequestDTO struct {
	Conflict   ConflictDTO `json:"conflict"`
	Resolution int         `json:"resolution"` // 1=KeepGlobal, 2=OverrideProject
}

// DoctorIssueDTO describes a single consistency problem.
type DoctorIssueDTO struct {
	Kind     string            `json:"kind"`
	Title    string            `json:"title"`
	Detail   string            `json:"detail"`
	HowToFix string            `json:"howToFix"`
	Fixable  bool              `json:"fixable"`
	FixData  map[string]string `json:"fixData"`
}

// DoctorReportDTO is the result of a health check run.
type DoctorReportDTO struct {
	Issues []DoctorIssueDTO `json:"issues"`
}

// RegisterProjectRequestDTO carries a manual project registration from the frontend.
type RegisterProjectRequestDTO struct {
	Path           string   `json:"path"`
	DetectedAgents []string `json:"detectedAgents"`
}

// --- Converters ---

func toSkillDTO(s domain.Skill) SkillDTO {
	return SkillDTO{
		ID:             s.ID,
		Name:           s.Name,
		Description:    s.Description,
		Path:           s.Path,
		Source:         string(s.Source),
		OwnerProjectID: s.OwnerProjectID,
		UpdatedAt:      s.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func toAggregatedSkillDTO(s usecase.AggregatedSkill) AggregatedSkillDTO {
	refs := make([]SkillProjectRef, len(s.Projects))
	for i, p := range s.Projects {
		refs[i] = SkillProjectRef{ID: p.ID, Name: p.Name, Path: p.Path, SkillPath: p.SkillPath}
	}
	return AggregatedSkillDTO{
		Name:         s.Name,
		Description:  s.Description,
		CategoryID:   s.CategoryID,
		CategoryName: s.CategoryName,
		IsGlobal:     s.IsGlobal,
		GlobalPath:   s.GlobalPath,
		Projects:     refs,
		UpdatedAt:    s.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func toCategoryDTO(c domain.Category) CategoryDTO {
	return CategoryDTO{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		CreatedAt:   c.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func toProjectCategoryLinkDTO(l domain.ProjectCategoryLink) ProjectCategoryLinkDTO {
	return ProjectCategoryLinkDTO{
		ProjectID:  l.ProjectID,
		CategoryID: l.CategoryID,
		Agent:      string(l.Agent),
		Category:   toCategoryDTO(l.Category),
	}
}

func toProjectDTO(p domain.Project) ProjectDTO {
	agents := make([]string, len(p.DetectedAgents))
	for i, a := range p.DetectedAgents {
		agents[i] = string(a)
	}
	return ProjectDTO{
		ID:             p.ID,
		Name:           p.Name,
		Path:           p.Path,
		DetectedAgents: agents,
		AddedAt:        p.AddedAt.UTC().Format(time.RFC3339),
	}
}

func toCandidateDTO(c usecase.ProjectCandidate) ProjectCandidateDTO {
	agents := make([]string, len(c.DetectedAgents))
	for i, a := range c.DetectedAgents {
		agents[i] = string(a)
	}
	return ProjectCandidateDTO{Name: c.Name, Path: c.Path, DetectedAgents: agents}
}

func toActivationDTO(a domain.Activation) ActivationDTO {
	pid := ""
	if a.ProjectID != nil {
		pid = *a.ProjectID
	}
	return ActivationDTO{
		ID:        a.ID,
		SkillID:   a.SkillID,
		Agent:     string(a.Agent),
		Scope:     string(a.Scope),
		ProjectID: pid,
		AppliedAt: a.AppliedAt.UTC().Format(time.RFC3339),
	}
}

func toConflictDTO(c domain.Conflict) ConflictDTO {
	dto := ConflictDTO{
		SkillID: c.SkillID,
		Agent:   string(c.Agent),
	}
	if c.GlobalActivation != nil {
		a := toActivationDTO(*c.GlobalActivation)
		dto.GlobalActivation = &a
	}
	if c.ProjectActivation != nil {
		a := toActivationDTO(*c.ProjectActivation)
		dto.ProjectActivation = &a
	}
	return dto
}

func toDoctorReportDTO(r usecase.DoctorReport) DoctorReportDTO {
	issues := make([]DoctorIssueDTO, len(r.Issues))
	for i, iss := range r.Issues {
		issues[i] = DoctorIssueDTO{
			Kind:     iss.Kind,
			Title:    iss.Title,
			Detail:   iss.Detail,
			HowToFix: iss.HowToFix,
			Fixable:  iss.Fixable,
			FixData:  iss.FixData,
		}
	}
	return DoctorReportDTO{Issues: issues}
}

func fromDoctorIssueDTO(dto DoctorIssueDTO) usecase.DoctorIssue {
	return usecase.DoctorIssue{
		Kind:     dto.Kind,
		Title:    dto.Title,
		Detail:   dto.Detail,
		HowToFix: dto.HowToFix,
		Fixable:  dto.Fixable,
		FixData:  dto.FixData,
	}
}
