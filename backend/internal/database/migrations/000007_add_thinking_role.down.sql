-- Revert to messages table without 'thinking' role
-- Note: This will DELETE all thinking messages

CREATE TABLE IF NOT EXISTS messages_old (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    role TEXT NOT NULL CHECK(role IN ('user', 'assistant')),
    content TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    FOREIGN KEY (session_id) REFERENCES sessions(session_id) ON DELETE CASCADE
);

-- Copy only user and assistant messages
INSERT OR IGNORE INTO messages_old (id, session_id, role, content, created_at)
SELECT id, session_id, role, content, created_at FROM messages WHERE role != 'thinking';

DROP TABLE IF EXISTS messages;

ALTER TABLE messages_old RENAME TO messages;

CREATE INDEX IF NOT EXISTS idx_messages_session_id ON messages(session_id);
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);
