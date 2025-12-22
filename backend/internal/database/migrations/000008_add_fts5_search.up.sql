-- Add FTS5 full-text search for messages
CREATE VIRTUAL TABLE IF NOT EXISTS messages_fts USING fts5(
    session_id UNINDEXED,
    role UNINDEXED,
    content,
    content=messages,
    content_rowid=id
);

-- Trigger for INSERT
CREATE TRIGGER IF NOT EXISTS messages_fts_ai AFTER INSERT ON messages BEGIN
    INSERT INTO messages_fts(rowid, session_id, role, content)
    VALUES (new.id, new.session_id, new.role, new.content);
END;

-- Trigger for DELETE
CREATE TRIGGER IF NOT EXISTS messages_fts_ad AFTER DELETE ON messages BEGIN
    INSERT INTO messages_fts(messages_fts, rowid, session_id, role, content)
    VALUES ('delete', old.id, old.session_id, old.role, old.content);
END;

-- Trigger for UPDATE
CREATE TRIGGER IF NOT EXISTS messages_fts_au AFTER UPDATE ON messages BEGIN
    INSERT INTO messages_fts(messages_fts, rowid, session_id, role, content)
    VALUES ('delete', old.id, old.session_id, old.role, old.content);
    INSERT INTO messages_fts(rowid, session_id, role, content)
    VALUES (new.id, new.session_id, new.role, new.content);
END;
