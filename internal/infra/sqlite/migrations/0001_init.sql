CREATE TABLE tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    description TEXT,
    priority TEXT NOT NULL DEFAULT 'medium' CHECK (priority IN ('low', 'medium', 'high')),
    created_at TEXT NOT NULL,
    completed_at TEXT,
    archived_at TEXT
);

CREATE INDEX idx_tasks_visible_order
    ON tasks (completed_at, archived_at, priority, created_at, id);

CREATE INDEX idx_tasks_created_at
    ON tasks (created_at);
