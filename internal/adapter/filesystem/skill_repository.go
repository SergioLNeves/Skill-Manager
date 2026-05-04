package filesystem

import (
	"context"
	"fmt"

	"github.com/spf13/afero"

	"skill-manager/internal/domain"
)

// SkillRepository reads global skills from one or more source directories.
// Each directory is scanned recursively; any subdirectory that contains a
// SKILL.md file is treated as one skill.
type SkillRepository struct {
	fs      afero.Fs
	sources []string
}

// NewSkillRepository creates a repository that scans the given source directories.
func NewSkillRepository(fs afero.Fs, sources ...string) *SkillRepository {
	return &SkillRepository{fs: fs, sources: sources}
}

// List returns all global skills found across all configured source directories.
func (r *SkillRepository) List(_ context.Context) ([]domain.Skill, error) {
	seen := make(map[string]bool)
	var skills []domain.Skill

	for _, src := range r.sources {
		found, err := scanSkillsDir(r.fs, src, false)
		if err != nil {
			continue // skip unreadable sources silently
		}
		for _, s := range found {
			if !seen[s.ID] {
				seen[s.ID] = true
				s.Source = domain.SkillSourceGlobal
				skills = append(skills, s)
			}
		}
	}
	return skills, nil
}

// GetByID returns the skill whose ID matches, or ErrSkillNotFound.
func (r *SkillRepository) GetByID(ctx context.Context, id string) (domain.Skill, error) {
	skills, err := r.List(ctx)
	if err != nil {
		return domain.Skill{}, err
	}
	for _, s := range skills {
		if s.ID == id {
			return s, nil
		}
	}
	return domain.Skill{}, fmt.Errorf("skill repository: %w", domain.ErrSkillNotFound)
}
