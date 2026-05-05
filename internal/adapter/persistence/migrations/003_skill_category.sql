ALTER TABLE skills ADD COLUMN category TEXT NOT NULL DEFAULT '';

PRAGMA user_version=3;
