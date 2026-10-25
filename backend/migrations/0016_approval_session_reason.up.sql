ALTER TABLE approval_requests ADD COLUMN session_id TEXT NOT NULL DEFAULT '';
ALTER TABLE approval_requests ADD COLUMN session_name TEXT NOT NULL DEFAULT '';
ALTER TABLE approval_requests ADD COLUMN reason TEXT NOT NULL DEFAULT '';
