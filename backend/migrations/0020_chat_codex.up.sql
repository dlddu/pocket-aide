ALTER TABLE chat_sessions ADD COLUMN platform_session_id TEXT NOT NULL DEFAULT '';
ALTER TABLE chat_sessions ADD COLUMN transcript_offset INTEGER NOT NULL DEFAULT 0;
