CREATE TABLE projects (
    id       TEXT PRIMARY KEY,
    name     TEXT NOT NULL,
    path     TEXT NOT NULL UNIQUE,
    added_at TIMESTAMP NOT NULL
);

CREATE TABLE project_agents (
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    agent      TEXT NOT NULL,
    PRIMARY KEY (project_id, agent)
);

CREATE TABLE activations (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    skill_id   TEXT NOT NULL,
    agent      TEXT NOT NULL,
    scope      TEXT NOT NULL CHECK (scope IN ('global', 'project')),
    project_id TEXT REFERENCES projects(id) ON DELETE CASCADE,
    applied_at TIMESTAMP NOT NULL,
    UNIQUE (skill_id, agent, scope, project_id)
);

CREATE INDEX idx_activations_lookup ON activations(skill_id, agent, project_id);
