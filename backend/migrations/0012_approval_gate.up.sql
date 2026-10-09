CREATE TABLE IF NOT EXISTS approval_caller_keys (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    key_hash TEXT NOT NULL UNIQUE,
    key_prefix TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    last_used_at INTEGER,
    revoked_at INTEGER
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_approval_caller_keys_active_name
    ON approval_caller_keys(user_id, name) WHERE revoked_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_approval_caller_keys_user_created
    ON approval_caller_keys(user_id, created_at);

CREATE TABLE IF NOT EXISTS approval_requests (
    id TEXT PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    caller_key_id INTEGER NOT NULL REFERENCES approval_caller_keys(id) ON DELETE CASCADE,
    external_id TEXT NOT NULL,
    context TEXT NOT NULL,
    requester_name TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'APPROVED', 'REJECTED')),
    decision_mode TEXT CHECK (decision_mode IN ('manual', 'auto')),
    created_at INTEGER NOT NULL,
    expires_at INTEGER,
    processed_at INTEGER,
    UNIQUE (caller_key_id, external_id)
);

CREATE INDEX IF NOT EXISTS idx_approval_requests_user_status_created
    ON approval_requests(user_id, status, created_at);
