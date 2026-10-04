CREATE TABLE IF NOT EXISTS user_notification_settings (
    user_id INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    enabled INTEGER NOT NULL DEFAULT 1 CHECK(enabled IN (0, 1)),
    outcomes TEXT NOT NULL DEFAULT 'both' CHECK(outcomes IN ('both', 'success', 'failure')),
    updated_at INTEGER NOT NULL DEFAULT (unixepoch())
);
