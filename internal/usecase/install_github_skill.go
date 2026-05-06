package usecase

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"skill-manager/internal/adapter/managed"
)

// SkillGitHubInstaller is satisfied by managed.SkillRepository.
type SkillGitHubInstaller interface {
	InstallGlobal(ctx context.Context, cacheDir, repo, ref, skillName string) (managed.InstallResult, error)
}

// SkillGitHubWriter writes GitHub-sourced skills to the cache DB.
type SkillGitHubWriter interface {
	UpsertSkill(ctx context.Context, name, description string, updatedAt time.Time) error
	UpsertGitHubLocation(ctx context.Context, skillName, repo, ref, subPath, path string) error
}

// InstallGitHubSkillRequest is the parsed install request.
type InstallGitHubSkillRequest struct {
	Repo      string
	Ref       string
	SkillName string
}

// InstallGitHubSkillResult lists the installed skills.
type InstallGitHubSkillResult struct {
	Skills []string
}

// InstallGitHubSkill installs skills from a GitHub repo into the global SQLite cache.
type InstallGitHubSkill struct {
	installer    SkillGitHubInstaller
	writer       SkillGitHubWriter
	globalCacheDir string
}

func NewInstallGitHubSkill(installer SkillGitHubInstaller, writer SkillGitHubWriter, globalCacheDir string) *InstallGitHubSkill {
	return &InstallGitHubSkill{installer: installer, writer: writer, globalCacheDir: globalCacheDir}
}

var reGitHubURL = regexp.MustCompile(`github\.com/([A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+?)(?:\.git)?(?:/|@(\S+))?(?:\s|$)`)
var reRepoSlug = regexp.MustCompile(`(?:^|\s)([A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+?)(?:@(\S+))?(?:\s|$)`)
var reRef = regexp.MustCompile(`--ref\s+(\S+)`)
var reSkill = regexp.MustCompile(`--skill\s+(\S+)`)

// ParseInstallInput extracts repo, ref, and optional skill name from a user-supplied string.
// Accepts formats:
//   - owner/repo
//   - owner/repo@ref
//   - https://github.com/owner/repo
//   - npx skills add owner/repo [--ref X] [--skill Y]
//   - skill-manager skills add owner/repo [flags]
func ParseInstallInput(input string) (InstallGitHubSkillRequest, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return InstallGitHubSkillRequest{}, fmt.Errorf("input vazio")
	}

	var req InstallGitHubSkillRequest

	// Extract --ref and --skill flags first.
	if m := reRef.FindStringSubmatch(input); m != nil {
		req.Ref = m[1]
	}
	if m := reSkill.FindStringSubmatch(input); m != nil {
		req.SkillName = m[1]
	}

	// Full GitHub URL: https://github.com/owner/repo or github.com/owner/repo
	if m := reGitHubURL.FindStringSubmatch(input); m != nil {
		req.Repo = m[1]
		if req.Ref == "" && m[2] != "" {
			req.Ref = m[2]
		}
		return req, nil
	}

	// Bare slug: owner/repo or owner/repo@ref (must not look like a domain)
	// Strip known CLI prefixes so we don't confuse "npx" as part of the slug.
	stripped := input
	for _, prefix := range []string{"npx skills add", "npx skill add", "skill-manager skills add"} {
		if idx := strings.Index(stripped, prefix); idx != -1 {
			stripped = stripped[idx+len(prefix):]
			break
		}
	}
	if m := reRepoSlug.FindStringSubmatch(" " + strings.TrimSpace(stripped)); m != nil {
		slug := m[1]
		// Reject slugs that look like domains (contain dots in the first segment).
		owner := strings.SplitN(slug, "/", 2)[0]
		if !strings.Contains(owner, ".") {
			req.Repo = slug
			if req.Ref == "" && m[2] != "" {
				req.Ref = m[2]
			}
			return req, nil
		}
	}

	return InstallGitHubSkillRequest{}, fmt.Errorf("não encontrei owner/repo em: %q — use owner/repo ou uma URL do GitHub", input)
}

func (uc *InstallGitHubSkill) Execute(ctx context.Context, input string) (InstallGitHubSkillResult, error) {
	req, err := ParseInstallInput(input)
	if err != nil {
		return InstallGitHubSkillResult{}, err
	}

	if err := os.MkdirAll(uc.globalCacheDir, 0o755); err != nil {
		return InstallGitHubSkillResult{}, fmt.Errorf("criar cache dir: %w", err)
	}

	result, err := uc.installer.InstallGlobal(ctx, uc.globalCacheDir, req.Repo, req.Ref, req.SkillName)
	if err != nil {
		return InstallGitHubSkillResult{}, fmt.Errorf("instalar skill: %w", err)
	}

	var names []string
	for _, skill := range result.Skills {
		desc := skill.Name // description not parsed from SKILL.md here — cache refresh will update
		if err := uc.writer.UpsertSkill(ctx, skill.Name, desc, skill.UpdatedAt); err != nil {
			return InstallGitHubSkillResult{}, fmt.Errorf("salvar skill %s: %w", skill.Name, err)
		}
		if err := uc.writer.UpsertGitHubLocation(ctx, skill.Name, skill.Repo, skill.Ref, skill.SubPath, skill.Path); err != nil {
			return InstallGitHubSkillResult{}, fmt.Errorf("salvar localização %s: %w", skill.Name, err)
		}
		names = append(names, skill.Name)
	}

	return InstallGitHubSkillResult{Skills: names}, nil
}
