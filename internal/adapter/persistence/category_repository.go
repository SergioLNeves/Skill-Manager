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

// CategoryRepository persists categories and project-category links in SQLite.
type CategoryRepository struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) Create(ctx context.Context, name, description string) (domain.Category, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO categories(name, description) VALUES(?,?)`, name, description)
	if err != nil {
		return domain.Category{}, fmt.Errorf("category repo: create %q: %w", name, err)
	}
	id, _ := res.LastInsertId()
	return r.GetByID(ctx, id)
}

func (r *CategoryRepository) Update(ctx context.Context, id int64, name, description string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE categories SET name=?, description=? WHERE id=?`, name, description, id)
	if err != nil {
		return fmt.Errorf("category repo: update %d: %w", id, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrCategoryNotFound
	}
	return nil
}

func (r *CategoryRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM categories WHERE id=?`, id)
	if err != nil {
		return fmt.Errorf("category repo: delete %d: %w", id, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrCategoryNotFound
	}
	return nil
}

func (r *CategoryRepository) List(ctx context.Context) ([]domain.Category, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, description, created_at FROM categories ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("category repo: list: %w", err)
	}
	defer rows.Close()
	var cats []domain.Category
	for rows.Next() {
		c, err := scanCategory(rows)
		if err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, rows.Err()
}

func (r *CategoryRepository) GetByID(ctx context.Context, id int64) (domain.Category, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, description, created_at FROM categories WHERE id=?`, id)
	c, err := scanCategory(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Category{}, domain.ErrCategoryNotFound
		}
		return domain.Category{}, fmt.Errorf("category repo: get %d: %w", id, err)
	}
	return c, nil
}

// AssignSkill sets (or clears) the category_id on the skill identified by name.
func (r *CategoryRepository) AssignSkill(ctx context.Context, skillName string, categoryID *int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE skills SET category_id=? WHERE name=?`, categoryID, skillName)
	if err != nil {
		return fmt.Errorf("category repo: assign skill %q: %w", skillName, err)
	}
	return nil
}

// ListSkillsInCategory returns the names of all skills assigned to the given category.
func (r *CategoryRepository) ListSkillsInCategory(ctx context.Context, categoryID int64) ([]string, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT name FROM skills WHERE category_id=? ORDER BY name`, categoryID)
	if err != nil {
		return nil, fmt.Errorf("category repo: list skills: %w", err)
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		names = append(names, n)
	}
	return names, rows.Err()
}

// LinkProject persists a project-category-agent association.
func (r *CategoryRepository) LinkProject(ctx context.Context, projectID string, categoryID int64, agent string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO project_categories(project_id, category_id, agent) VALUES(?,?,?)
		 ON CONFLICT DO NOTHING`,
		projectID, categoryID, agent)
	if err != nil {
		return fmt.Errorf("category repo: link project: %w", err)
	}
	return nil
}

// UnlinkProject removes a project-category-agent association.
func (r *CategoryRepository) UnlinkProject(ctx context.Context, projectID string, categoryID int64, agent string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM project_categories WHERE project_id=? AND category_id=? AND agent=?`,
		projectID, categoryID, agent)
	if err != nil {
		return fmt.Errorf("category repo: unlink project: %w", err)
	}
	return nil
}

// ListProjectCategories returns all category links for the given project.
func (r *CategoryRepository) ListProjectCategories(ctx context.Context, projectID string) ([]domain.ProjectCategoryLink, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT pc.project_id, pc.category_id, pc.agent,
		       c.id, c.name, c.description, c.created_at
		FROM project_categories pc
		JOIN categories c ON c.id = pc.category_id
		WHERE pc.project_id = ?
		ORDER BY c.name, pc.agent`,
		projectID)
	if err != nil {
		return nil, fmt.Errorf("category repo: list project categories: %w", err)
	}
	defer rows.Close()
	return scanProjectCategoryLinks(rows)
}

// ListProjectsLinkedToCategory returns all project links for a given category.
func (r *CategoryRepository) ListProjectsLinkedToCategory(ctx context.Context, categoryID int64) ([]domain.ProjectCategoryLink, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT pc.project_id, pc.category_id, pc.agent,
		       c.id, c.name, c.description, c.created_at
		FROM project_categories pc
		JOIN categories c ON c.id = pc.category_id
		WHERE pc.category_id = ?
		ORDER BY pc.project_id, pc.agent`,
		categoryID)
	if err != nil {
		return nil, fmt.Errorf("category repo: list projects for category: %w", err)
	}
	defer rows.Close()
	return scanProjectCategoryLinks(rows)
}

// GetSkillCategoryAndLinks returns the current category_id of a skill and all
// project-links for that category (used by AssignSkillCategory propagation).
func (r *CategoryRepository) GetSkillPreviousCategoryLinks(ctx context.Context, skillName string) ([]domain.ProjectCategoryLink, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT pc.project_id, pc.category_id, pc.agent,
		       c.id, c.name, c.description, c.created_at
		FROM skills s
		JOIN project_categories pc ON pc.category_id = s.category_id
		JOIN categories c ON c.id = pc.category_id
		WHERE s.name = ?`,
		skillName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanProjectCategoryLinks(rows)
}

// GetSkillCategoryLinks returns all project-category links for a given category_id.
// This is the same as ListProjectsLinkedToCategory but used internally.
func (r *CategoryRepository) GetCategoryLinks(ctx context.Context, categoryID int64) ([]domain.ProjectCategoryLink, error) {
	return r.ListProjectsLinkedToCategory(ctx, categoryID)
}

// ListCategorySkillsForProject returns AggregatedSkill-like data for each skill in
// a category that the cache knows about, for use in copy propagation.
func (r *CategoryRepository) ListCategorySkillPaths(ctx context.Context, categoryID int64) ([]usecase.CategorySkillPath, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT s.name, COALESCE(
			(SELECT l.path FROM skill_locations l WHERE l.skill_name = s.name AND l.source = 'github' LIMIT 1),
			(SELECT l.path FROM skill_locations l WHERE l.skill_name = s.name LIMIT 1)
		)
		FROM skills s
		WHERE s.category_id = ?`,
		categoryID)
	if err != nil {
		return nil, fmt.Errorf("category repo: list skill paths: %w", err)
	}
	defer rows.Close()
	var results []usecase.CategorySkillPath
	for rows.Next() {
		var name string
		var path sql.NullString
		if err := rows.Scan(&name, &path); err != nil {
			return nil, err
		}
		if path.Valid {
			results = append(results, usecase.CategorySkillPath{SkillName: name, Path: path.String})
		}
	}
	return results, rows.Err()
}

func scanCategory(s rowScanner) (domain.Category, error) {
	var c domain.Category
	var createdAt string
	if err := s.Scan(&c.ID, &c.Name, &c.Description, &createdAt); err != nil {
		return domain.Category{}, err
	}
	t, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		// fallback: SQLite may return datetime without T
		t, err = time.Parse("2006-01-02 15:04:05", createdAt)
		if err != nil {
			return domain.Category{}, fmt.Errorf("parse created_at %q: %w", createdAt, err)
		}
	}
	c.CreatedAt = t
	return c, nil
}

func scanProjectCategoryLinks(rows *sql.Rows) ([]domain.ProjectCategoryLink, error) {
	var links []domain.ProjectCategoryLink
	for rows.Next() {
		var link domain.ProjectCategoryLink
		var agent string
		var createdAt string
		err := rows.Scan(
			&link.ProjectID, &link.CategoryID, &agent,
			&link.Category.ID, &link.Category.Name, &link.Category.Description, &createdAt,
		)
		if err != nil {
			return nil, err
		}
		link.Agent = domain.Agent(agent)
		t, err := time.Parse(time.RFC3339, createdAt)
		if err != nil {
			t, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		}
		link.Category.CreatedAt = t
		links = append(links, link)
	}
	return links, rows.Err()
}
