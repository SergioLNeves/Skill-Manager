package filesystem

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/spf13/afero"

	"skill-manager/internal/domain"
)

// ProjectSkillRepository discovers skills that live inside a project directory.
// It scans two locations:
//   - <project>/skills/         — user-authored skills, scanned normally.
//   - <project>/.claude/skills/ — scanned with symlink filtering, because
//     the ClaudeAdapter places activation symlinks there.
type ProjectSkillRepository struct {
	fs afero.Fs
}

// NewProjectSkillRepository creates a repository backed by the given filesystem.
func NewProjectSkillRepository(fs afero.Fs) *ProjectSkillRepository {
	return &ProjectSkillRepository{fs: fs}
}

// ListByProject returns all skills found inside the given project directory.
func (r *ProjectSkillRepository) ListByProject(_ context.Context, project domain.Project) ([]domain.Skill, error) {
	seen := make(map[string]bool)
	var skills []domain.Skill

	roots := []struct {
		path        string
		skipSymlink bool
	}{
		{filepath.Join(project.Path, "skills"), false},
		{filepath.Join(project.Path, ".claude", "skills"), true},
	}

	for _, root := range roots {
		found, err := scanSkillsDir(r.fs, root.path, root.skipSymlink)
		if err != nil {
			continue
		}
		for _, s := range found {
			if seen[s.ID] {
				continue
			}
			seen[s.ID] = true
			s.Source = domain.SkillSourceProject
			s.OwnerProjectID = project.ID
			skills = append(skills, s)
		}
	}
	return skills, nil
}

// GetByID returns a project skill whose ID matches, or ErrSkillNotFound.
func (r *ProjectSkillRepository) GetByID(ctx context.Context, id string, project domain.Project) (domain.Skill, error) {
	skills, err := r.ListByProject(ctx, project)
	if err != nil {
		return domain.Skill{}, err
	}
	for _, s := range skills {
		if s.ID == id {
			return s, nil
		}
	}
	return domain.Skill{}, fmt.Errorf("project skill repository: %w", domain.ErrSkillNotFound)
}
