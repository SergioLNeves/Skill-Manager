package usecase

import (
	"context"
	"fmt"
)

// SkillCacheReader is implemented by the SQLite skill cache repository.
type SkillCacheReader interface {
	ListAggregated(ctx context.Context) ([]AggregatedSkill, error)
}

// ListAllSkills returns skills aggregated from the SQLite cache (refreshed on each call).
type ListAllSkills struct {
	refresh *RefreshSkillCache
	cache   SkillCacheReader
}

func NewListAllSkills(refresh *RefreshSkillCache, cache SkillCacheReader) *ListAllSkills {
	return &ListAllSkills{refresh: refresh, cache: cache}
}

func (uc *ListAllSkills) Execute(ctx context.Context) ([]AggregatedSkill, error) {
	if err := uc.refresh.Execute(ctx); err != nil {
		return nil, fmt.Errorf("list all skills: refresh cache: %w", err)
	}
	skills, err := uc.cache.ListAggregated(ctx)
	if err != nil {
		return nil, fmt.Errorf("list all skills: read cache: %w", err)
	}
	return skills, nil
}
