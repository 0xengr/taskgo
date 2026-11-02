CREATE TABLE projects (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    name_key TEXT NOT NULL UNIQUE,
    repo_path TEXT UNIQUE,
    created_at TEXT NOT NULL,
    archived_at TEXT
);

ALTER TABLE tasks ADD COLUMN project_id INTEGER REFERENCES projects(id);

CREATE INDEX idx_tasks_project ON tasks (project_id);
