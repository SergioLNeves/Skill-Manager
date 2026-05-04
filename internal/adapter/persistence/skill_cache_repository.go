package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"skill-manager/internal/usecase"
)

// SkillCacheRepository persists aggregated skill data to SQLite.
type SkillCacheRepository struct {
	db          *sql.DB
	projectRepo *ProjectRepository
}

func NewSkillCacheRepository(db *sql.DB, projectRepo *ProjectRepository) *SkillCacheRepository {
	return &SkillCacheRepository{db: db, projectRepo: projectRepo}
}

func (r *SkillCacheRepository) UpsertSkill(ctx context.Context, name, description string, updatedAt time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO skills(name, description, updated_at) VALUES(?,?,?)
		 ON CONFLICT(name) DO UPDATE SET description=excluded.description, updated_at=excluded.updated_at`,
		name, description, updatedAt.UTC(),
	)
	return err
}

func (r *SkillCacheRepository) UpsertLocation(ctx context.Context, skillName, source, projectID, path string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO skill_locations(skill_name, source, project_id, path) VALUES(?,?,?,?)
		 ON CONFLICT(skill_name, source, project_id) DO UPDATE SET path=excluded.path`,
		skillName, source, projectID, path,
	)
	return err
}

func (r *SkillCacheRepository) DeleteGlobalLocations(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM skill_locations WHERE source='global'`)
	return err
}

func (r *SkillCacheRepository) DeleteLocationsForProject(ctx context.Context, projectID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM skill_locations WHERE source='project' AND project_id=?`, projectID)
	return err
}

func (r *SkillCacheRepository) DeleteSkill(ctx context.Context, name string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM skills WHERE name=?`, name)
	return err
}

// PruneOrphanSkills removes skills that have no remaining locations.
func (r *SkillCacheRepository) PruneOrphanSkills(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM skills WHERE name NOT IN (SELECT DISTINCT skill_name FROM skill_locations)`)
	return err
}

// ListAggregated returns all skills with their locations joined, satisfying usecase.SkillCacheReader.
func (r *SkillCacheRepository) ListAggregated(ctx context.Context) ([]usecase.AggregatedSkill, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT s.name, s.description, s.updated_at,
		       l.source, l.project_id, l.path
		FROM skills s
		JOIN skill_locations l ON l.skill_name = s.name
		ORDER BY s.name, l.source, l.project_id
	`)
	if err != nil {
		return nil, fmt.Errorf("skill cache: list aggregated: %w", err)
	}
	defer rows.Close()

	byName := make(map[string]*usecase.AggregatedSkill)
	var order []string

	for rows.Next() {
		var name, description, source, projectID, path string
		var updatedAt time.Time

		if err := rows.Scan(&name, &description, &updatedAt, &source, &projectID, &path); err != nil {
			return nil, err
		}

		agg, exists := byName[name]
		if !exists {
			agg = &usecase.AggregatedSkill{
				Name:        name,
				Description: description,
				UpdatedAt:   updatedAt,
			}
			byName[name] = agg
			order = append(order, name)
		}

		if source == "global" {
			agg.IsGlobal = true
			agg.GlobalPath = path
		} else if projectID != "" {
			p, err := r.projectRepo.GetByID(ctx, projectID)
			pName := projectID
			pPath := ""
			if err == nil {
				pName = p.Name
				pPath = p.Path
			}
			agg.Projects = append(agg.Projects, usecase.SkillProjectRef{
				ID:   projectID,
				Name: pName,
				Path: pPath,
			})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	result := make([]usecase.AggregatedSkill, 0, len(order))
	for _, n := range order {
		result = append(result, *byName[n])
	}
	return result, nil
}
