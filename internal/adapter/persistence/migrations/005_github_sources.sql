-- Extend skill_locations to support GitHub-sourced skills.
-- Adds repo, ref (resolved SHA), and sub_path columns.
-- Updates source CHECK to allow 'github' alongside 'project'.
-- Migrates existing 'global' rows to 'github'.

ALTER TABLE skill_locations RENAME TO skill_locations_old;

CREATE TABLE skill_locations (
    skill_name TEXT NOT NULL REFERENCES skills(name) ON DELETE CASCADE,
    source     TEXT NOT NULL CHECK (source IN ('github', 'project')),
    project_id TEXT NOT NULL DEFAULT '',
    repo       TEXT NOT NULL DEFAULT '',
    ref        TEXT NOT NULL DEFAULT '',
    sub_path   TEXT NOT NULL DEFAULT '',
    path       TEXT NOT NULL,
    PRIMARY KEY (skill_name, source, project_id, repo)
);

CREATE INDEX IF NOT EXISTS idx_skill_locations_project ON skill_locations(project_id);

INSERT INTO skill_locations (skill_name, source, project_id, path)
SELECT skill_name,
       CASE source WHEN 'global' THEN 'github' ELSE 'project' END,
       project_id,
       path
FROM skill_locations_old;

DROP TABLE skill_locations_old;
