CREATE TABLE IF NOT EXISTS categories (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE skills ADD COLUMN category_id INTEGER REFERENCES categories(id) ON DELETE SET NULL;

INSERT INTO categories(name)
SELECT DISTINCT TRIM(category) FROM skills
WHERE category IS NOT NULL AND TRIM(category) != ''
ON CONFLICT(name) DO NOTHING;

UPDATE skills
SET category_id = (SELECT id FROM categories WHERE categories.name = TRIM(skills.category))
WHERE category IS NOT NULL AND TRIM(category) != '';

ALTER TABLE skills DROP COLUMN category;

CREATE TABLE IF NOT EXISTS project_categories (
    project_id  TEXT    NOT NULL REFERENCES projects(id)   ON DELETE CASCADE,
    category_id INTEGER NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    agent       TEXT    NOT NULL,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (project_id, category_id, agent)
);

CREATE INDEX IF NOT EXISTS idx_project_categories_category ON project_categories(category_id);

PRAGMA user_version=4;
