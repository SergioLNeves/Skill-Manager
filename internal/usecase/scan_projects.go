package usecase

import (
	"context"
	"fmt"
)

// ScanProjects suggests project candidates from given workspace roots.
// Persisting the results requires a separate RegisterProject call per candidate.
type ScanProjects struct {
	scanner ProjectScanner
}

func NewScanProjects(scanner ProjectScanner) *ScanProjects {
	return &ScanProjects{scanner: scanner}
}

func (uc *ScanProjects) Execute(ctx context.Context, roots []string) ([]ProjectCandidate, error) {
	candidates, err := uc.scanner.Scan(ctx, roots)
	if err != nil {
		return nil, fmt.Errorf("scan projects: %w", err)
	}
	return candidates, nil
}
