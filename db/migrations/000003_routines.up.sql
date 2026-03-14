CREATE TABLE IF NOT EXISTS routines (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id         INTEGER NOT NULL,
    name            TEXT    NOT NULL,
    interval_days   INTEGER NOT NULL,
    last_done_at    TEXT    NOT NULL,
    next_due_date   TEXT    NOT NULL,
    note            TEXT    DEFAULT '',
    notify_enabled  INTEGER DEFAULT 0,
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
