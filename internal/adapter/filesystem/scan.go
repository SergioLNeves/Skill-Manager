package filesystem

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"

	"skill-manager/internal/domain"
)

const (
	skillFile        = "SKILL.md"
	frontmatterDelim = "---"
)

// skipDirs are directory names that are never scanned for skills.
var skipDirs = map[string]bool{
	"node_modules": true,
	"vendor":       true,
	".git":         true,
	"dist":         true,
	"build":        true,
	".cache":       true,
	"__pycache__":  true,
}

// scanSkillsDir walks root recursively, collecting every directory that
// contains a SKILL.md file. When skipSymlinks is true, symlinked entries in
// root are ignored (used to skip activation symlinks in .claude/skills/).
func scanSkillsDir(fs afero.Fs, root string, skipSymlinks bool) ([]domain.Skill, error) {
	var skills []domain.Skill

	var walk func(dir string, isRoot bool) error
	walk = func(dir string, isRoot bool) error {
		entries, err := afero.ReadDir(fs, dir)
		if err != nil {
			if isRoot {
				return nil // missing root is not an error
			}
			return nil
		}

		// Check if this directory itself contains a SKILL.md.
		for _, e := range entries {
			if !e.IsDir() && e.Name() == skillFile {
				skill, readErr := readSkill(fs, dir)
				if readErr == nil {
					skills = append(skills, skill)
				}
				// Don't recurse into a skill directory itself.
				return nil
			}
		}

		// Recurse into subdirectories.
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			name := e.Name()
			if strings.HasPrefix(name, ".") || skipDirs[name] {
				continue
			}
			childPath := filepath.Join(dir, name)
			if isRoot && skipSymlinks && isSymlink(childPath) {
				continue
			}
			_ = walk(childPath, false)
		}
		return nil
	}

	if err := walk(root, true); err != nil {
		return nil, fmt.Errorf("scan skills dir %s: %w", root, err)
	}
	return skills, nil
}

func readSkill(fs afero.Fs, dir string) (domain.Skill, error) {
	mdPath := filepath.Join(dir, skillFile)

	f, err := fs.Open(mdPath)
	if err != nil {
		return domain.Skill{}, fmt.Errorf("open %s: %w", mdPath, err)
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		return domain.Skill{}, fmt.Errorf("read %s: %w", mdPath, err)
	}

	info, err := fs.Stat(dir)
	if err != nil {
		return domain.Skill{}, fmt.Errorf("stat %s: %w", dir, err)
	}

	meta := parseFrontmatter(content)
	displayName := meta["name"]
	if displayName == "" {
		displayName = filepath.Base(dir)
	}
	desc := meta["description"]
	if desc == "" {
		desc = displayName
	}

	return domain.Skill{
		ID:          stableID(dir),
		Name:        displayName,
		Path:        dir,
		Description: desc,
		UpdatedAt:   info.ModTime(),
	}, nil
}

// stableID returns a deterministic ID from the skill's canonical path.
// Canonicalizing via EvalSymlinks ensures the same skill reached through
// different symlinks (e.g. .claude/skills/foo → skills/foo) maps to one ID.
func stableID(path string) string {
	if canonical, err := filepath.EvalSymlinks(path); err == nil {
		path = canonical
	}
	sum := sha256.Sum256([]byte(path))
	return fmt.Sprintf("%x", sum[:8])
}

// parseFrontmatter extracts key-value pairs from a YAML frontmatter block.
func parseFrontmatter(content []byte) map[string]string {
	s := string(bytes.TrimSpace(content))
	if !strings.HasPrefix(s, frontmatterDelim) {
		return map[string]string{}
	}

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

// isSymlink returns true when path is a symbolic link on the real filesystem.
func isSymlink(path string) bool {
	fi, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeSymlink != 0
}

// now is a seam used in tests.
var now = time.Now
