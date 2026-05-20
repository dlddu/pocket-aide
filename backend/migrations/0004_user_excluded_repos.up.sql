CREATE TABLE IF NOT EXISTS user_excluded_repos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    repo_full_name TEXT NOT NULL,
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    UNIQUE(user_id, repo_full_name)
);

CREATE INDEX IF NOT EXISTS idx_user_excluded_repos_repo ON user_excluded_repos(repo_full_name);
