package binding

import (
	"context"
	"fmt"

	"skill-manager/internal/domain"
	"skill-manager/internal/usecase"
)

// ActivationBinding exposes skill activation operations to the Wails frontend.
type ActivationBinding struct {
	activate        *usecase.ActivateSkill
	deactivate      *usecase.DeactivateSkill
	resolveConflict *usecase.ResolveConflict
	activations     usecase.ActivationRepository
}

func NewActivationBinding(
	activate *usecase.ActivateSkill,
	deactivate *usecase.DeactivateSkill,
	resolve *usecase.ResolveConflict,
	activations usecase.ActivationRepository,
) *ActivationBinding {
	return &ActivationBinding{
		activate:        activate,
		deactivate:      deactivate,
		resolveConflict: resolve,
		activations:     activations,
	}
}

// List returns activations matching the given filter.
func (b *ActivationBinding) List(ctx context.Context, f ActivationFilterDTO) ([]ActivationDTO, error) {
	filter := usecase.ActivationFilter{
		SkillID:   f.SkillID,
		Agent:     domain.Agent(f.Agent),
		Scope:     domain.Scope(f.Scope),
		ProjectID: f.ProjectID,
	}
	activations, err := b.activations.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("activation binding: list: %w", err)
	}
	dtos := make([]ActivationDTO, len(activations))
	for i, a := range activations {
		dtos[i] = toActivationDTO(a)
	}
	return dtos, nil
}

// Activate activates a skill. Returns a conflict when one is detected.
func (b *ActivationBinding) Activate(ctx context.Context, req ActivateRequestDTO) (ActivateResultDTO, error) {
	result, err := b.activate.Execute(ctx, usecase.ActivateSkillRequest{
		SkillID:   req.SkillID,
		Agent:     domain.Agent(req.Agent),
		Scope:     domain.Scope(req.Scope),
		ProjectID: req.ProjectID,
	})
	if err != nil {
		return ActivateResultDTO{}, fmt.Errorf("activation binding: activate: %w", err)
	}

	dto := ActivateResultDTO{}
	if result.Conflict != nil {
		c := toConflictDTO(*result.Conflict)
		dto.Conflict = &c
	}
	if result.Activation != nil {
		a := toActivationDTO(*result.Activation)
		dto.Activation = &a
	}
	return dto, nil
}

// Deactivate removes an activation by ID.
func (b *ActivationBinding) Deactivate(ctx context.Context, id int64) error {
	if err := b.deactivate.Execute(ctx, id); err != nil {
		return fmt.Errorf("activation binding: deactivate %d: %w", id, err)
	}
	return nil
}

// ResolveConflict applies the user's chosen resolution.
func (b *ActivationBinding) ResolveConflict(ctx context.Context, req ResolveConflictRequestDTO) error {
	var globalAct, projectAct *domain.Activation
	if req.Conflict.GlobalActivation != nil {
		a := fromActivationDTO(*req.Conflict.GlobalActivation)
		globalAct = &a
	}
	if req.Conflict.ProjectActivation != nil {
		a := fromActivationDTO(*req.Conflict.ProjectActivation)
		projectAct = &a
	}
	conflict := domain.Conflict{
		SkillID:           req.Conflict.SkillID,
		Agent:             domain.Agent(req.Conflict.Agent),
		GlobalActivation:  globalAct,
		ProjectActivation: projectAct,
	}
	err := b.resolveConflict.Execute(ctx, usecase.ResolveConflictRequest{
		Conflict:   conflict,
		Resolution: domain.ConflictResolution(req.Resolution),
	})
	if err != nil {
		return fmt.Errorf("activation binding: resolve conflict: %w", err)
	}
	return nil
}

// DoctorBinding exposes the doctor health check to the Wails frontend.
type DoctorBinding struct {
	doctor *usecase.Doctor
}

func NewDoctorBinding(doctor *usecase.Doctor) *DoctorBinding {
	return &DoctorBinding{doctor: doctor}
}

// Run executes the health check and returns any detected issues.
func (b *DoctorBinding) Run(ctx context.Context) (DoctorReportDTO, error) {
	report, err := b.doctor.Execute(ctx)
	if err != nil {
		return DoctorReportDTO{}, fmt.Errorf("doctor binding: %w", err)
	}
	return toDoctorReportDTO(report), nil
}

// fromActivationDTO converts a DTO back to a domain.Activation for conflict resolution.
func fromActivationDTO(dto ActivationDTO) domain.Activation {
	a := domain.Activation{
		ID:      dto.ID,
		SkillID: dto.SkillID,
		Agent:   domain.Agent(dto.Agent),
		Scope:   domain.Scope(dto.Scope),
	}
	if dto.ProjectID != "" {
		a.ProjectID = &dto.ProjectID
	}
	return a
}
