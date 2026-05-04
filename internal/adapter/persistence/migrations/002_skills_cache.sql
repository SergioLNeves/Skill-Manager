CREATE TABLE skills (
    name        TEXT PRIMARY KEY,
    description TEXT NOT NULL,
    updated_at  DATETIME NOT NULL
);

CREATE TABLE skill_locations (
    skill_name TEXT NOT NULL REFERENCES skills(name) ON DELETE CASCADE,
    source     TEXT NOT NULL CHECK (source IN ('global', 'project')),
    project_id TEXT,
    path       TEXT NOT NULL,
    PRIMARY KEY (skill_name, source, COALESCE(project_id, ''))
);

CREATE INDEX idx_skill_locations_project ON skill_locations(project_id);
