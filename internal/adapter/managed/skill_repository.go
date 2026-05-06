package managed

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	githubfetcher "skill-manager/internal/adapter/github"
	"skill-manager/internal/domain"
)

// SkillRepository manages skills installed from GitHub repos.
// Skills are installed per-project under <projectDir>/.skills/cache/.
type SkillRepository struct {
	fetcher *githubfetcher.Fetcher
}

// NewSkillRepository creates a ManagedSkillRepository backed by the given fetcher.
func NewSkillRepository(fetcher *githubfetcher.Fetcher) *SkillRepository {
	return &SkillRepository{fetcher: fetcher}
}

// InstallRequest describes a skill to install.
type InstallRequest struct {
	ProjectDir string // the project root (where skills-lock.json lives)
	Repo       string // "owner/repo"
	Ref        string // branch, tag, or SHA; defaults to "main"
	SkillName  string // specific skill to install; empty means install all from repo
}

// InstallResult contains the installed skills.
type InstallResult struct {
	Skills []domain.Skill
}

// InstallGlobal downloads the GitHub repo to a global cache dir and returns the found skills.
// Unlike Install it does NOT write a skills-lock.json.
func (r *SkillRepository) InstallGlobal(ctx context.Context, cacheDir, repo, ref, skillName string) (InstallResult, error) {
	if ref == "" {
		ref = "main"
	}

	result, err := r.fetcher.Fetch(ctx, cacheDir, repo, ref)
	if err != nil {
		return InstallResult{}, fmt.Errorf("managed repo: fetch: %w", err)
	}

	skillDirs, err := findSkillDirs(result.ExtractDir, skillName)
	if err != nil {
		return InstallResult{}, fmt.Errorf("managed repo: find skills: %w", err)
	}
	if len(skillDirs) == 0 {
		return InstallResult{}, fmt.Errorf("managed repo: no skill named %q found in %s", skillName, repo)
	}

	var skills []domain.Skill
	for _, dir := range skillDirs {
		name := filepath.Base(dir)
		subPath, _ := filepath.Rel(result.ExtractDir, dir)

		info, _ := os.Stat(dir)
		var updatedAt time.Time
		if info != nil {
			updatedAt = info.ModTime()
		}
		skills = append(skills, domain.Skill{
			Name:      name,
			Path:      dir,
			Source:    domain.SkillSourceGitHub,
			UpdatedAt: updatedAt,
			Repo:      repo,
			SubPath:   filepath.ToSlash(subPath),
			Ref:       result.SHA,
		})
	}

	return InstallResult{Skills: skills}, nil
}

// Install downloads the GitHub repo and registers matching skills in the lock file.
func (r *SkillRepository) Install(ctx context.Context, req InstallRequest) (InstallResult, error) {
	if req.Ref == "" {
		req.Ref = "main"
	}
	cacheDir := filepath.Join(req.ProjectDir, ".skills", "cache")

	result, err := r.fetcher.Fetch(ctx, cacheDir, req.Repo, req.Ref)
	if err != nil {
		return InstallResult{}, fmt.Errorf("managed repo: fetch: %w", err)
	}

	lf, err := ReadLockFile(req.ProjectDir)
	if err != nil {
		return InstallResult{}, fmt.Errorf("managed repo: read lockfile: %w", err)
	}

	skillDirs, err := findSkillDirs(result.ExtractDir, req.SkillName)
	if err != nil {
		return InstallResult{}, fmt.Errorf("managed repo: find skills: %w", err)
	}
	if len(skillDirs) == 0 {
		return InstallResult{}, fmt.Errorf("managed repo: no skill named %q found in %s", req.SkillName, req.Repo)
	}

	var skills []domain.Skill
	for _, dir := range skillDirs {
		name := filepath.Base(dir)
		skillMD := filepath.Join(dir, "SKILL.md")

		hash, err := githubfetcher.HashFile(skillMD)
		if err != nil {
			continue
		}

		// skillPath relative to extractDir (e.g. "skills/(code-quality)/go-coding-standards/SKILL.md")
		relPath, _ := filepath.Rel(result.ExtractDir, skillMD)
		subPath, _ := filepath.Rel(result.ExtractDir, dir)

		entry := LockEntry{
			Source:       req.Repo,
			SourceType:   "github",
			SkillPath:    filepath.ToSlash(relPath),
			ComputedHash: hash,
			Ref:          result.SHA,
		}
		if err := lf.Add(name, entry); err != nil {
			return InstallResult{}, fmt.Errorf("managed repo: update lockfile for %s: %w", name, err)
		}

		info, _ := os.Stat(dir)
		var updatedAt time.Time
		if info != nil {
			updatedAt = info.ModTime()
		}
		skills = append(skills, domain.Skill{
			Name:      name,
			Path:      dir,
			Source:    domain.SkillSourceGitHub,
			UpdatedAt: updatedAt,
			Repo:      req.Repo,
			SubPath:   filepath.ToSlash(subPath),
			Ref:       result.SHA,
		})
	}

	return InstallResult{Skills: skills}, nil
}

// ListFromLockFile reads the lock file and returns all installed skills with their local paths.
func ListFromLockFile(projectDir string) ([]domain.Skill, error) {
	lf, err := ReadLockFile(projectDir)
	if err != nil {
		return nil, err
	}
	cacheDir := filepath.Join(projectDir, ".skills", "cache")

	var skills []domain.Skill
	for name, entry := range lf.Skills {
		subPath := filepath.Dir(filepath.FromSlash(entry.SkillPath))
		localPath := filepath.Join(cacheDir, entry.Source, entry.Ref, subPath)

		info, _ := os.Stat(localPath)
		var updatedAt time.Time
		if info != nil {
			updatedAt = info.ModTime()
		}

		skills = append(skills, domain.Skill{
			Name:      name,
			Path:      localPath,
			Source:    domain.SkillSourceGitHub,
			UpdatedAt: updatedAt,
			Repo:      entry.Source,
			SubPath:   subPath,
			Ref:       entry.Ref,
		})
	}
	return skills, nil
}

// findSkillDirs searches extractDir for directories containing SKILL.md.
// If skillName is non-empty, only the matching directory is returned.
func findSkillDirs(extractDir, skillName string) ([]string, error) {
	var result []string

	err := filepath.Walk(extractDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() || info.Name() != "SKILL.md" {
			if !info.IsDir() && info.Name() == "SKILL.md" {
				dir := filepath.Dir(path)
				name := filepath.Base(dir)
				if skillName == "" || name == skillName {
					result = append(result, dir)
				}
			}
			return nil
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// filepath.Walk visits files, not directories named SKILL.md — re-walk correctly
	result = nil
	_ = filepath.Walk(extractDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if info.Name() == "SKILL.md" {
			dir := filepath.Dir(path)
			// Skip if inside a hidden or vendor directory
			rel, _ := filepath.Rel(extractDir, dir)
			if strings.Contains(rel, "..") {
				return nil
			}
			name := filepath.Base(dir)
			if skillName == "" || name == skillName {
				result = append(result, dir)
			}
		}
		return nil
	})

	return result, nil
}
