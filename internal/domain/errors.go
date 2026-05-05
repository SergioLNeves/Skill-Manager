package domain

import "errors"

var (
	ErrSkillNotFound   = errors.New("domain: skill not found")
	ErrProjectNotFound = errors.New("domain: project not found")
	ErrConflict        = errors.New("domain: activation conflict")
	ErrInvalidScope    = errors.New("domain: invalid scope")
	ErrActivationNotFound = errors.New("domain: activation not found")
	ErrCategoryNotFound   = errors.New("domain: category not found")
)
