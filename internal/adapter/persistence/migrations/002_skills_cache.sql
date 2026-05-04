CREATE TABLE IF NOT EXISTS skills (
    name        TEXT PRIMARY KEY,
    description TEXT NOT NULL,
    updated_at  DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS skill_locations (
    skill_name TEXT NOT NULL REFERENCES skills(name) ON DELETE CASCADE,
    source     TEXT NOT NULL CHECK (source IN ('global', 'project')),
    project_id TEXT NOT NULL DEFAULT '',
    path       TEXT NOT NULL,
    PRIMARY KEY (skill_name, source, project_id)
);

CREATE INDEX IF NOT EXISTS idx_skill_locations_project ON skill_locations(project_id);
