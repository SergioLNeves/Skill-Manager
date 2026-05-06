package managed

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const lockFileName = "skills-lock.json"

// LockEntry describes one installed skill in the lock file.
type LockEntry struct {
	Source       string `json:"source"`       // "owner/repo"
	SourceType   string `json:"sourceType"`   // always "github"
	SkillPath    string `json:"skillPath"`    // path inside repo to SKILL.md
	ComputedHash string `json:"computedHash"` // SHA-256 of SKILL.md content
	Ref          string `json:"ref,omitempty"` // resolved commit SHA
}

// LockFile is the in-memory representation of skills-lock.json.
type LockFile struct {
	Version int                  `json:"version"`
	Skills  map[string]LockEntry `json:"skills"`
	path    string
}

// ReadLockFile loads (or creates) the lock file at <dir>/skills-lock.json.
func ReadLockFile(dir string) (*LockFile, error) {
	p := filepath.Join(dir, lockFileName)
	data, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return &LockFile{Version: 1, Skills: map[string]LockEntry{}, path: p}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("lockfile: read %s: %w", p, err)
	}
	var lf LockFile
	if err := json.Unmarshal(data, &lf); err != nil {
		return nil, fmt.Errorf("lockfile: parse %s: %w", p, err)
	}
	if lf.Skills == nil {
		lf.Skills = map[string]LockEntry{}
	}
	lf.path = p
	return &lf, nil
}

// Add inserts or updates a skill entry and writes the file.
func (lf *LockFile) Add(name string, entry LockEntry) error {
	lf.Skills[name] = entry
	return lf.write()
}

// Remove deletes a skill entry and writes the file.
func (lf *LockFile) Remove(name string) error {
	delete(lf.Skills, name)
	return lf.write()
}

func (lf *LockFile) write() error {
	data, err := json.MarshalIndent(lf, "", "  ")
	if err != nil {
		return fmt.Errorf("lockfile: marshal: %w", err)
	}
	if err := os.WriteFile(lf.path, data, 0o644); err != nil {
		return fmt.Errorf("lockfile: write %s: %w", lf.path, err)
	}
	return nil
}

// Path returns the absolute path to this lock file.
func (lf *LockFile) Path() string { return lf.path }
