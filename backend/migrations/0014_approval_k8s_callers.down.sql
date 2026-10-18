DROP INDEX IF EXISTS idx_approval_caller_keys_active_service_account;
ALTER TABLE approval_caller_keys DROP COLUMN service_account;
ALTER TABLE approval_caller_keys DROP COLUMN kind;
