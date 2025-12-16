package models

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

// Session represents a conversation session with Claude Code
type Session struct {
	ID              int       `json:"id"`
	SessionID       string    `json:"session_id"`        // Internal session UUID
	ClaudeSessionID string    `json:"claude_session_id"` // Claude Code CLI session ID for --resume
	Title           string    `json:"title"`             // Auto-generated title from first message
	CreatedAt       time.Time `json:"created_at"`
	LastActivity    time.Time `json:"last_activity"`
}

// Message represents a single message in a conversation
type Message struct {
	ID        int       `json:"id"`
	SessionID string    `json:"session_id"` // References Session.SessionID
	Role      string    `json:"role"`       // "user" or "assistant"
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// DB wraps the SQLite database connection
type DB struct {
	conn *sql.DB
}

// InitDB initializes the SQLite database and creates tables if they don't exist
func InitDB(dbPath string) (*DB, error) {
	log.Printf("Initializing database at: %s", dbPath)

	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DB{conn: conn}

	// Create tables
	if err := db.createTables(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	log.Println("Database initialized successfully")
	return db, nil
}

// createTables creates the necessary database tables
func (db *DB) createTables() error {
	sessionsTable := `
	CREATE TABLE IF NOT EXISTS sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id TEXT UNIQUE NOT NULL,
		claude_session_id TEXT DEFAULT '',
		title TEXT DEFAULT '',
		created_at DATETIME NOT NULL,
		last_activity DATETIME NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_sessions_session_id ON sessions(session_id);
	CREATE INDEX IF NOT EXISTS idx_sessions_last_activity ON sessions(last_activity DESC);
	`

	// Migration: add title column if it doesn't exist
	alterTableTitle := `
	ALTER TABLE sessions ADD COLUMN title TEXT DEFAULT '';
	`

	// Migration: add claude_session_id column if it doesn't exist
	alterTableClaudeSession := `
	ALTER TABLE sessions ADD COLUMN claude_session_id TEXT DEFAULT '';
	`

	messagesTable := `
	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id TEXT NOT NULL,
		role TEXT NOT NULL CHECK(role IN ('user', 'assistant')),
		content TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		FOREIGN KEY (session_id) REFERENCES sessions(session_id) ON DELETE CASCADE
	);
	CREATE INDEX IF NOT EXISTS idx_messages_session_id ON messages(session_id);
	CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);
	`

	// Execute table creation queries
	if _, err := db.conn.Exec(sessionsTable); err != nil {
		return fmt.Errorf("failed to create sessions table: %w", err)
	}

	if _, err := db.conn.Exec(messagesTable); err != nil {
		return fmt.Errorf("failed to create messages table: %w", err)
	}

	// Run migrations (ignore errors if columns already exist)
	db.conn.Exec(alterTableTitle)
	db.conn.Exec(alterTableClaudeSession)

	return nil
}

// CreateSession creates a new session in the database
func (db *DB) CreateSession(sessionID string) (*Session, error) {
	now := time.Now()

	query := `
	INSERT INTO sessions (session_id, claude_session_id, title, created_at, last_activity)
	VALUES (?, '', '', ?, ?)
	`

	result, err := db.conn.Exec(query, sessionID, now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	log.Printf("Created new session: %s (ID: %d)", sessionID, id)

	return &Session{
		ID:              int(id),
		SessionID:       sessionID,
		ClaudeSessionID: "",
		Title:           "",
		CreatedAt:       now,
		LastActivity:    now,
	}, nil
}

// GetSession retrieves a session by its session ID
func (db *DB) GetSession(sessionID string) (*Session, error) {
	query := `
	SELECT id, session_id, COALESCE(claude_session_id, ''), title, created_at, last_activity
	FROM sessions
	WHERE session_id = ?
	`

	var session Session
	err := db.conn.QueryRow(query, sessionID).Scan(
		&session.ID,
		&session.SessionID,
		&session.ClaudeSessionID,
		&session.Title,
		&session.CreatedAt,
		&session.LastActivity,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Session not found
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &session, nil
}

// UpdateSessionActivity updates the last activity timestamp for a session
func (db *DB) UpdateSessionActivity(sessionID string) error {
	query := `
	UPDATE sessions
	SET last_activity = ?
	WHERE session_id = ?
	`

	result, err := db.conn.Exec(query, time.Now(), sessionID)
	if err != nil {
		return fmt.Errorf("failed to update session activity: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	return nil
}

// SaveMessage saves a message to the database
func (db *DB) SaveMessage(sessionID, role, content string) (*Message, error) {
	now := time.Now()

	query := `
	INSERT INTO messages (session_id, role, content, created_at)
	VALUES (?, ?, ?, ?)
	`

	result, err := db.conn.Exec(query, sessionID, role, content, now)
	if err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	log.Printf("Saved %s message for session %s (ID: %d)", role, sessionID, id)

	return &Message{
		ID:        int(id),
		SessionID: sessionID,
		Role:      role,
		Content:   content,
		CreatedAt: now,
	}, nil
}

// GetMessages retrieves all messages for a session, ordered by creation time
func (db *DB) GetMessages(sessionID string) ([]*Message, error) {
	query := `
	SELECT id, session_id, role, content, created_at
	FROM messages
	WHERE session_id = ?
	ORDER BY created_at ASC
	`

	rows, err := db.conn.Query(query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		var msg Message
		err := rows.Scan(
			&msg.ID,
			&msg.SessionID,
			&msg.Role,
			&msg.Content,
			&msg.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, &msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating messages: %w", err)
	}

	return messages, nil
}

// ListSessions retrieves all sessions ordered by last activity (most recent first)
func (db *DB) ListSessions() ([]*Session, error) {
	query := `
	SELECT id, session_id, COALESCE(claude_session_id, ''), title, created_at, last_activity
	FROM sessions
	ORDER BY last_activity DESC
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		var session Session
		err := rows.Scan(
			&session.ID,
			&session.SessionID,
			&session.ClaudeSessionID,
			&session.Title,
			&session.CreatedAt,
			&session.LastActivity,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, &session)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sessions: %w", err)
	}

	return sessions, nil
}

// UpdateSessionTitle updates the title of a session
func (db *DB) UpdateSessionTitle(sessionID, title string) error {
	query := `
	UPDATE sessions
	SET title = ?
	WHERE session_id = ?
	`

	result, err := db.conn.Exec(query, title, sessionID)
	if err != nil {
		return fmt.Errorf("failed to update session title: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	return nil
}

// UpdateClaudeSessionID updates the Claude CLI session ID for a session
func (db *DB) UpdateClaudeSessionID(sessionID, claudeSessionID string) error {
	query := `
	UPDATE sessions
	SET claude_session_id = ?
	WHERE session_id = ?
	`

	result, err := db.conn.Exec(query, claudeSessionID, sessionID)
	if err != nil {
		return fmt.Errorf("failed to update claude session id: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	log.Printf("Updated Claude session ID for %s: %s", sessionID, claudeSessionID)
	return nil
}

// DeleteSession deletes a session and all its messages
func (db *DB) DeleteSession(sessionID string) error {
	// Delete messages first (foreign key)
	_, err := db.conn.Exec("DELETE FROM messages WHERE session_id = ?", sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete messages: %w", err)
	}

	// Delete session
	result, err := db.conn.Exec("DELETE FROM sessions WHERE session_id = ?", sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	log.Printf("Deleted session: %s", sessionID)
	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	if db.conn != nil {
		log.Println("Closing database connection")
		return db.conn.Close()
	}
	return nil
}

// GetConnection returns the underlying database connection
func (db *DB) GetConnection() *sql.DB {
	return db.conn
}
