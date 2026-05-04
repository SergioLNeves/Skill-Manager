package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"skill-manager/internal/domain"
	"skill-manager/internal/usecase"
)

// ActivationRepository persists skill activations in SQLite.
type ActivationRepository struct {
	db *sql.DB
}

func NewActivationRepository(db *sql.DB) *ActivationRepository {
	return &ActivationRepository{db: db}
}

func (r *ActivationRepository) List(ctx context.Context, f usecase.ActivationFilter) ([]domain.Activation, error) {
	q := `SELECT id, skill_id, agent, scope, project_id, applied_at FROM activations WHERE 1=1`
	var args []any

	if f.SkillID != "" {
		q += " AND skill_id = ?"
		args = append(args, f.SkillID)
	}
	if f.Agent != "" {
		q += " AND agent = ?"
		args = append(args, string(f.Agent))
	}
	if f.Scope != "" {
		q += " AND scope = ?"
		args = append(args, string(f.Scope))
	}
	if f.ProjectID != "" {
		q += " AND project_id = ?"
		args = append(args, f.ProjectID)
	}
	q += " ORDER BY id"

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("activation repo: list: %w", err)
	}
	defer rows.Close()

	var activations []domain.Activation
	for rows.Next() {
		a, scanErr := scanActivation(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("activation repo: scan: %w", scanErr)
		}
		activations = append(activations, a)
	}
	return activations, rows.Err()
}

func (r *ActivationRepository) Save(ctx context.Context, a domain.Activation) (domain.Activation, error) {
	var projectID *string
	if a.ProjectID != nil {
		projectID = a.ProjectID
	}

	res, err := r.db.ExecContext(ctx,
		`INSERT INTO activations (skill_id, agent, scope, project_id, applied_at)
		 VALUES (?, ?, ?, ?, ?)`,
		a.SkillID, string(a.Agent), string(a.Scope), projectID,
		a.AppliedAt.UTC().Format(time.RFC3339))
	if err != nil {
		return domain.Activation{}, fmt.Errorf("activation repo: save: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return domain.Activation{}, fmt.Errorf("activation repo: last insert id: %w", err)
	}
	a.ID = id
	return a, nil
}

func (r *ActivationRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM activations WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("activation repo: delete %d: %w", id, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrActivationNotFound
	}
	return nil
}

// FindConflict returns a Conflict when the same (skillID, agent) pair has both
// a global activation and a project-scoped activation for the given projectID.
func (r *ActivationRepository) FindConflict(ctx context.Context, skillID string, agent domain.Agent, projectID string) (*domain.Conflict, error) {
	global, err := r.findActivation(ctx, skillID, string(agent), string(domain.ScopeGlobal), "")
	if err != nil {
		return nil, err
	}
	project, err := r.findActivation(ctx, skillID, string(agent), string(domain.ScopeProject), projectID)
	if err != nil {
		return nil, err
	}
	if global == nil || project == nil {
		return nil, nil
	}
	return &domain.Conflict{
		SkillID:           skillID,
		Agent:             agent,
		GlobalActivation:  global,
		ProjectActivation: project,
	}, nil
}

func (r *ActivationRepository) findActivation(ctx context.Context, skillID, agent, scope, projectID string) (*domain.Activation, error) {
	var row *sql.Row
	if scope == string(domain.ScopeGlobal) {
		row = r.db.QueryRowContext(ctx,
			`SELECT id, skill_id, agent, scope, project_id, applied_at
			 FROM activations
			 WHERE skill_id=? AND agent=? AND scope='global' AND project_id IS NULL`,
			skillID, agent)
	} else {
		row = r.db.QueryRowContext(ctx,
			`SELECT id, skill_id, agent, scope, project_id, applied_at
			 FROM activations
			 WHERE skill_id=? AND agent=? AND scope='project' AND project_id=?`,
			skillID, agent, projectID)
	}

	a, err := scanActivationRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("activation repo: find %s/%s/%s: %w", skillID, agent, scope, err)
	}
	return &a, nil
}

type activationScanner interface {
	Scan(dest ...any) error
}

func scanActivation(s activationScanner) (domain.Activation, error) {
	var a domain.Activation
	var projectID sql.NullString
	var appliedAt string
	if err := s.Scan(&a.ID, &a.SkillID, &a.Agent, &a.Scope, &projectID, &appliedAt); err != nil {
		return domain.Activation{}, err
	}
	if projectID.Valid {
		a.ProjectID = &projectID.String
	}
	t, err := time.Parse(time.RFC3339, appliedAt)
	if err != nil {
		return domain.Activation{}, fmt.Errorf("parse applied_at %q: %w", appliedAt, err)
	}
	a.AppliedAt = t
	return a, nil
}

func scanActivationRow(row *sql.Row) (domain.Activation, error) {
	var a domain.Activation
	var projectID sql.NullString
	var appliedAt string
	if err := row.Scan(&a.ID, &a.SkillID, &a.Agent, &a.Scope, &projectID, &appliedAt); err != nil {
		return domain.Activation{}, err
	}
	if projectID.Valid {
		a.ProjectID = &projectID.String
	}
	t, err := time.Parse(time.RFC3339, appliedAt)
	if err != nil {
		return domain.Activation{}, fmt.Errorf("parse applied_at %q: %w", appliedAt, err)
	}
	a.AppliedAt = t
	return a, nil
}
