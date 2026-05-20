CREATE TABLE IF NOT EXISTS notification_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    repo_full_name TEXT NOT NULL,
    pr_number INTEGER,
    pr_title TEXT,
    pr_url TEXT,
    commit_url TEXT,
    run_url TEXT,
    workflow_name TEXT NOT NULL DEFAULT '',
    head_branch TEXT NOT NULL DEFAULT '',
    conclusion TEXT NOT NULL,
    acknowledged_at INTEGER,
    created_at INTEGER NOT NULL DEFAULT (unixepoch())
);

CREATE INDEX IF NOT EXISTS idx_notif_history_user_created
    ON notification_history(user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_notif_history_user_acked
    ON notification_history(user_id, acknowledged_at);
