ALTER TABLE approval_caller_keys ADD COLUMN kind TEXT NOT NULL DEFAULT 'key' CHECK (kind IN ('key', 'k8s_sa'));
ALTER TABLE approval_caller_keys ADD COLUMN service_account TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS idx_approval_caller_keys_active_service_account
    ON approval_caller_keys(service_account) WHERE revoked_at IS NULL AND kind = 'k8s_sa';
