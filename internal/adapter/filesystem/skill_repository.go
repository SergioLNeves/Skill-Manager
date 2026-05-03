package filesystem

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"

	"skill-manager/internal/domain"
)

const (
	skillFile    = "SKILL.md"
	frontmatterDelim = "---"
)

// SkillRepository reads skills from the central repository directory.
// Each subdirectory containing a SKILL.md file is treated as one skill.
type SkillRepository struct {
	fs   afero.Fs
	root string
}

// NewSkillRepository creates a repository rooted at the given path.
func NewSkillRepository(fs afero.Fs, root string) *SkillRepository {
	return &SkillRepository{fs: fs, root: root}
}

// List returns all skills found under the root directory.
func (r *SkillRepository) List(_ context.Context) ([]domain.Skill, error) {
	entries, err := afero.ReadDir(r.fs, r.root)
	if err != nil {
		return nil, fmt.Errorf("skill repository: read dir %s: %w", r.root, err)
	}

	skills := make([]domain.Skill, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		skill, err := r.readSkill(e.Name())
		if err != nil {
			continue // skip unreadable or malformed skill dirs
		}
		skills = append(skills, skill)
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
	return domain.Skill{}, domain.ErrSkillNotFound
}

func (r *SkillRepository) readSkill(name string) (domain.Skill, error) {
	dir := filepath.Join(r.root, name)
	mdPath := filepath.Join(dir, skillFile)

	f, err := r.fs.Open(mdPath)
	if err != nil {
		return domain.Skill{}, fmt.Errorf("open %s: %w", mdPath, err)
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		return domain.Skill{}, fmt.Errorf("read %s: %w", mdPath, err)
	}

	info, err := r.fs.Stat(dir)
	if err != nil {
		return domain.Skill{}, fmt.Errorf("stat %s: %w", dir, err)
	}

	meta := parseFrontmatter(content)
	desc := meta["description"]
	if desc == "" {
		desc = meta["name"]
	}

	return domain.Skill{
		ID:          stableID(name),
		Name:        name,
		Path:        dir,
		Description: desc,
		UpdatedAt:   info.ModTime(),
	}, nil
}

// stableID returns a deterministic ID from the skill directory name.
func stableID(name string) string {
	sum := sha256.Sum256([]byte(name))
	return fmt.Sprintf("%x", sum[:8])
}

// parseFrontmatter extracts key-value pairs from a YAML frontmatter block.
// Returns an empty map when no valid frontmatter is found.
func parseFrontmatter(content []byte) map[string]string {
	s := string(bytes.TrimSpace(content))
	if !strings.HasPrefix(s, frontmatterDelim) {
		return map[string]string{}
	}

	// Strip opening delimiter
	s = strings.TrimPrefix(s, frontmatterDelim)
	end := strings.Index(s, "\n"+frontmatterDelim)
	if end == -1 {
		return map[string]string{}
	}

	block := strings.TrimSpace(s[:end])
	var out map[string]string
	if err := yaml.Unmarshal([]byte(block), &out); err != nil {
		return map[string]string{}
	}
	if out == nil {
		return map[string]string{}
	}
	return out
}

// now is a seam used in tests.
var now = time.Now
