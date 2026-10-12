CREATE TABLE IF NOT EXISTS user_approval_settings (
    user_id INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    auto_response TEXT NOT NULL DEFAULT 'NONE' CHECK (auto_response IN ('NONE', 'AUTO_APPROVE', 'AUTO_REJECT')),
    updated_at INTEGER NOT NULL DEFAULT (unixepoch())
);
