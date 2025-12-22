-- Add claude_session_id column to sessions
ALTER TABLE sessions ADD COLUMN claude_session_id TEXT DEFAULT '';
