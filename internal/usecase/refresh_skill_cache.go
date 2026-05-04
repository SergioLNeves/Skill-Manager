package usecase

import (
	"context"
	"fmt"
	"path/filepath"
	"time"
)

// SkillCacheWriter is implemented by the SQLite skill cache repository.
type SkillCacheWriter interface {
	UpsertSkill(ctx context.Context, name, description string, updatedAt time.Time) error
	UpsertLocation(ctx context.Context, skillName, source, projectID, path string) error
	DeleteGlobalLocations(ctx context.Context) error
	DeleteLocationsForProject(ctx context.Context, projectID string) error
	PruneOrphanSkills(ctx context.Context) error
}

// RefreshSkillCache rebuilds the SQLite skill cache from the filesystem.
type RefreshSkillCache struct {
	skills        SkillRepository
	projects      ProjectRepository
	projectSkills ProjectSkillRepository
	cache         SkillCacheWriter
}

func NewRefreshSkillCache(
	skills SkillRepository,
	projects ProjectRepository,
	projectSkills ProjectSkillRepository,
	cache SkillCacheWriter,
) *RefreshSkillCache {
	return &RefreshSkillCache{
		skills:        skills,
		projects:      projects,
		projectSkills: projectSkills,
		cache:         cache,
	}
}

func (uc *RefreshSkillCache) Execute(ctx context.Context) error {
	// Refresh global skills.
	if err := uc.cache.DeleteGlobalLocations(ctx); err != nil {
		return fmt.Errorf("refresh cache: clear globals: %w", err)
	}
	globals, err := uc.skills.List(ctx)
	if err != nil {
		return fmt.Errorf("refresh cache: list globals: %w", err)
	}
	for _, s := range globals {
		name := filepath.Base(s.Path)
		if err := uc.cache.UpsertSkill(ctx, name, s.Description, s.UpdatedAt); err != nil {
			return fmt.Errorf("refresh cache: upsert global skill %s: %w", name, err)
		}
		if err := uc.cache.UpsertLocation(ctx, name, "global", "", s.Path); err != nil {
			return fmt.Errorf("refresh cache: upsert global location %s: %w", name, err)
		}
	}

	// Refresh per-project skills.
	projects, err := uc.projects.List(ctx)
	if err != nil {
		return fmt.Errorf("refresh cache: list projects: %w", err)
	}
	for _, p := range projects {
		if err := uc.cache.DeleteLocationsForProject(ctx, p.ID); err != nil {
			return fmt.Errorf("refresh cache: clear project %s: %w", p.ID, err)
		}
		skills, err := uc.projectSkills.ListByProject(ctx, p)
		if err != nil {
			continue // skip unreachable directories
		}
		for _, s := range skills {
			name := filepath.Base(s.Path)
			if err := uc.cache.UpsertSkill(ctx, name, s.Description, s.UpdatedAt); err != nil {
				return fmt.Errorf("refresh cache: upsert project skill %s: %w", name, err)
			}
			if err := uc.cache.UpsertLocation(ctx, name, "project", p.ID, s.Path); err != nil {
				return fmt.Errorf("refresh cache: upsert project location %s: %w", name, err)
			}
		}
	}

	return uc.cache.PruneOrphanSkills(ctx)
}
