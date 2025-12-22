-- Add tool_calls table
CREATE TABLE IF NOT EXISTS tool_calls (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    tool_use_id TEXT UNIQUE NOT NULL,
    tool_name TEXT NOT NULL,
    input TEXT NOT NULL DEFAULT '{}',
    output TEXT DEFAULT '',
    status TEXT NOT NULL CHECK(status IN ('running', 'success', 'error')),
    created_at DATETIME NOT NULL,
    completed_at DATETIME,
    FOREIGN KEY (session_id) REFERENCES sessions(session_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_tool_calls_session_id ON tool_calls(session_id);
CREATE INDEX IF NOT EXISTS idx_tool_calls_tool_use_id ON tool_calls(tool_use_id);
