package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"skill-manager/internal/domain"
)

// ProjectRepository persists projects in SQLite.
type ProjectRepository struct {
	db *sql.DB
}

func NewProjectRepository(db *sql.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) List(ctx context.Context) ([]domain.Project, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, path, added_at FROM projects ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("project repo: list: %w", err)
	}

	// Collect all rows before closing so we can query agents on the same connection.
	var projects []domain.Project
	for rows.Next() {
		p, scanErr := scanProject(rows)
		if scanErr != nil {
			rows.Close()
			return nil, fmt.Errorf("project repo: scan: %w", scanErr)
		}
		projects = append(projects, p)
	}
	rows.Close()
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("project repo: list rows: %w", err)
	}

	for i := range projects {
		agents, agentErr := r.loadAgents(ctx, projects[i].ID)
		if agentErr != nil {
			return nil, agentErr
		}
		projects[i].DetectedAgents = agents
	}
	return projects, nil
}

func (r *ProjectRepository) GetByID(ctx context.Context, id string) (domain.Project, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, path, added_at FROM projects WHERE id = ?`, id)

	p, err := scanProjectRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Project{}, domain.ErrProjectNotFound
		}
		return domain.Project{}, fmt.Errorf("project repo: get %s: %w", id, err)
	}

	agents, err := r.loadAgents(ctx, p.ID)
	if err != nil {
		return domain.Project{}, err
	}
	p.DetectedAgents = agents
	return p, nil
}

func (r *ProjectRepository) Save(ctx context.Context, p domain.Project) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("project repo: begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	_, err = tx.ExecContext(ctx,
		`INSERT INTO projects (id, name, path, added_at)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET name=excluded.name, path=excluded.path`,
		p.ID, p.Name, p.Path, p.AddedAt.UTC().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("project repo: upsert %s: %w", p.ID, err)
	}

	if _, err = tx.ExecContext(ctx,
		`DELETE FROM project_agents WHERE project_id = ?`, p.ID); err != nil {
		return fmt.Errorf("project repo: clear agents %s: %w", p.ID, err)
	}

	for _, agent := range p.DetectedAgents {
		if _, err = tx.ExecContext(ctx,
			`INSERT INTO project_agents (project_id, agent) VALUES (?, ?)`,
			p.ID, string(agent)); err != nil {
			return fmt.Errorf("project repo: insert agent %s for %s: %w", agent, p.ID, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("project repo: commit %s: %w", p.ID, err)
	}
	return nil
}

func (r *ProjectRepository) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM projects WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("project repo: delete %s: %w", id, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrProjectNotFound
	}
	return nil
}

func (r *ProjectRepository) loadAgents(ctx context.Context, projectID string) ([]domain.Agent, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT agent FROM project_agents WHERE project_id = ? ORDER BY agent`, projectID)
	if err != nil {
		return nil, fmt.Errorf("project repo: load agents %s: %w", projectID, err)
	}
	defer rows.Close()

	var agents []domain.Agent
	for rows.Next() {
		var a string
		if err = rows.Scan(&a); err != nil {
			return nil, fmt.Errorf("project repo: scan agent: %w", err)
		}
		agents = append(agents, domain.Agent(a))
	}
	return agents, rows.Err()
}

type projectScanner interface {
	Scan(dest ...any) error
}

func scanProject(s projectScanner) (domain.Project, error) {
	var p domain.Project
	var addedAt string
	if err := s.Scan(&p.ID, &p.Name, &p.Path, &addedAt); err != nil {
		return domain.Project{}, err
	}
	t, err := time.Parse(time.RFC3339, addedAt)
	if err != nil {
		return domain.Project{}, fmt.Errorf("parse added_at %q: %w", addedAt, err)
	}
	p.AddedAt = t
	return p, nil
}

func scanProjectRow(row *sql.Row) (domain.Project, error) {
	var p domain.Project
	var addedAt string
	if err := row.Scan(&p.ID, &p.Name, &p.Path, &addedAt); err != nil {
		return domain.Project{}, err
	}
	t, err := time.Parse(time.RFC3339, addedAt)
	if err != nil {
		return domain.Project{}, fmt.Errorf("parse added_at %q: %w", addedAt, err)
	}
	p.AddedAt = t
	return p, nil
}
