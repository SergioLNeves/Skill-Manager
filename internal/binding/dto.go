package binding

import (
	"time"

	"skill-manager/internal/domain"
	"skill-manager/internal/usecase"
)

// SkillDTO is the frontend representation of a skill.
type SkillDTO struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	Path             string `json:"path"`
	Source           string `json:"source"`           // "global" or "project"
	OwnerProjectID   string `json:"ownerProjectId"`   // non-empty when Source == "project"
	OwnerProjectName string `json:"ownerProjectName"` // display name of owning project
	UpdatedAt        string `json:"updatedAt"`
}

// AggregatedSkillDTO is the frontend representation of a deduplicated skill with all locations.
type AggregatedSkillDTO struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	IsGlobal    bool                `json:"isGlobal"`
	GlobalPath  string              `json:"globalPath"`
	Projects    []SkillProjectRef   `json:"projects"`
	UpdatedAt   string              `json:"updatedAt"`
}

// SkillProjectRef is a lightweight project reference inside an aggregated skill.
type SkillProjectRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
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
	Kind    string `json:"kind"`
	Detail  string `json:"detail"`
	Fixable bool   `json:"fixable"`
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
		refs[i] = SkillProjectRef{ID: p.ID, Name: p.Name, Path: p.Path}
	}
	return AggregatedSkillDTO{
		Name:        s.Name,
		Description: s.Description,
		IsGlobal:    s.IsGlobal,
		GlobalPath:  s.GlobalPath,
		Projects:    refs,
		UpdatedAt:   s.UpdatedAt.UTC().Format(time.RFC3339),
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
		issues[i] = DoctorIssueDTO{Kind: iss.Kind, Detail: iss.Detail, Fixable: iss.Fixable}
	}
	return DoctorReportDTO{Issues: issues}
}
