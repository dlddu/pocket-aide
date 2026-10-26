CREATE TABLE IF NOT EXISTS external_notifications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    caller_key_id INTEGER NOT NULL REFERENCES approval_caller_keys(id) ON DELETE CASCADE,
    requester TEXT NOT NULL DEFAULT '',
    title TEXT NOT NULL,
    body TEXT NOT NULL DEFAULT '',
    devices INTEGER NOT NULL DEFAULT 0,
    sent INTEGER NOT NULL DEFAULT 0,
    read_at INTEGER,
    created_at INTEGER NOT NULL DEFAULT (unixepoch())
);

CREATE INDEX IF NOT EXISTS idx_external_notifications_user_id
    ON external_notifications(user_id, id DESC);

CREATE INDEX IF NOT EXISTS idx_external_notifications_user_read
    ON external_notifications(user_id, read_at);
