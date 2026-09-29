-- PRD-3 AC1: personal and work todos are separate data collections, not one
-- table with an area tag. The two tables share a shape but never a row, so no
-- query can leak one area's items into the other's list or search.
CREATE TABLE IF NOT EXISTS personal_todos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    memo TEXT NOT NULL DEFAULT '',
    due_date TEXT,
    priority TEXT CHECK(priority IN ('high', 'normal', 'low')),
    completed_at INTEGER,
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    updated_at INTEGER NOT NULL DEFAULT (unixepoch())
);

CREATE INDEX IF NOT EXISTS idx_personal_todos_user_id ON personal_todos(user_id);

CREATE TABLE IF NOT EXISTS work_todos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    memo TEXT NOT NULL DEFAULT '',
    due_date TEXT,
    priority TEXT CHECK(priority IN ('high', 'normal', 'low')),
    completed_at INTEGER,
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    updated_at INTEGER NOT NULL DEFAULT (unixepoch())
);

CREATE INDEX IF NOT EXISTS idx_work_todos_user_id ON work_todos(user_id);
