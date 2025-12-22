-- Add 'thinking' role support to messages table
-- SQLite doesn't support ALTER TABLE to modify CHECK constraints
-- We need to recreate the table

-- Step 1: Create new table with updated constraint
CREATE TABLE IF NOT EXISTS messages_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    role TEXT NOT NULL CHECK(role IN ('user', 'assistant', 'thinking')),
    content TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    FOREIGN KEY (session_id) REFERENCES sessions(session_id) ON DELETE CASCADE
);

-- Step 2: Copy existing data
INSERT OR IGNORE INTO messages_new (id, session_id, role, content, created_at)
SELECT id, session_id, role, content, created_at FROM messages;

-- Step 3: Drop old table
DROP TABLE IF EXISTS messages;

-- Step 4: Rename new table
ALTER TABLE messages_new RENAME TO messages;

-- Step 5: Recreate indexes
CREATE INDEX IF NOT EXISTS idx_messages_session_id ON messages(session_id);
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);
