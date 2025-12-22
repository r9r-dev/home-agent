package models

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "modernc.org/sqlite"
)

// DB wraps the SQLite database connection
type DB struct {
	conn *sql.DB
}

// InitDB initializes the SQLite database and creates tables if they don't exist
func InitDB(dbPath string) (*DB, error) {
	log.Printf("Initializing database at: %s", dbPath)

	// Add SQLite pragmas for better concurrent access
	dsn := dbPath + "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool for concurrent access
	conn.SetMaxOpenConns(1) // SQLite works best with single writer
	conn.SetMaxIdleConns(1)

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

	// Migration: add model column if it doesn't exist (default to haiku)
	alterTableModel := `
	ALTER TABLE sessions ADD COLUMN model TEXT DEFAULT 'haiku';
	`

	messagesTable := `
	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id TEXT NOT NULL,
		role TEXT NOT NULL CHECK(role IN ('user', 'assistant', 'thinking')),
		content TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		FOREIGN KEY (session_id) REFERENCES sessions(session_id) ON DELETE CASCADE
	);
	CREATE INDEX IF NOT EXISTS idx_messages_session_id ON messages(session_id);
	CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);
	`

	settingsTable := `
	CREATE TABLE IF NOT EXISTS settings (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL DEFAULT '',
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	`

	memoryTable := `
	CREATE TABLE IF NOT EXISTS memory (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		enabled INTEGER DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	toolCallsTable := `
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
	`

	machinesTable := `
	CREATE TABLE IF NOT EXISTS machines (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT DEFAULT '',
		host TEXT NOT NULL,
		port INTEGER DEFAULT 22,
		username TEXT NOT NULL,
		auth_type TEXT NOT NULL CHECK(auth_type IN ('password', 'key')),
		auth_value TEXT NOT NULL,
		status TEXT DEFAULT 'untested' CHECK(status IN ('untested', 'online', 'offline')),
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_machines_name ON machines(name);
	`

	// Execute table creation queries
	if _, err := db.conn.Exec(sessionsTable); err != nil {
		return fmt.Errorf("failed to create sessions table: %w", err)
	}

	if _, err := db.conn.Exec(messagesTable); err != nil {
		return fmt.Errorf("failed to create messages table: %w", err)
	}

	if _, err := db.conn.Exec(settingsTable); err != nil {
		return fmt.Errorf("failed to create settings table: %w", err)
	}

	if _, err := db.conn.Exec(memoryTable); err != nil {
		return fmt.Errorf("failed to create memory table: %w", err)
	}

	if _, err := db.conn.Exec(toolCallsTable); err != nil {
		return fmt.Errorf("failed to create tool_calls table: %w", err)
	}

	if _, err := db.conn.Exec(machinesTable); err != nil {
		return fmt.Errorf("failed to create machines table: %w", err)
	}

	// Run migrations (ignore errors if columns already exist)
	db.conn.Exec(alterTableTitle)
	db.conn.Exec(alterTableClaudeSession)
	db.conn.Exec(alterTableModel)

	// Migration: update messages table to allow 'thinking' role
	// SQLite doesn't support ALTER TABLE to modify CHECK constraints, so we need to recreate the table
	if err := db.migrateMessagesTableForThinking(); err != nil {
		log.Printf("Warning: failed to migrate messages table for thinking role: %v", err)
	}

	// FTS5 virtual table for full-text search on messages
	fts5Table := `
	CREATE VIRTUAL TABLE IF NOT EXISTS messages_fts USING fts5(
		session_id UNINDEXED,
		role UNINDEXED,
		content,
		content=messages,
		content_rowid=id
	);
	`

	if _, err := db.conn.Exec(fts5Table); err != nil {
		return fmt.Errorf("failed to create FTS5 table: %w", err)
	}

	// Triggers to keep FTS table in sync with messages table
	fts5Triggers := []string{
		// Trigger for INSERT
		`CREATE TRIGGER IF NOT EXISTS messages_fts_ai AFTER INSERT ON messages BEGIN
			INSERT INTO messages_fts(rowid, session_id, role, content)
			VALUES (new.id, new.session_id, new.role, new.content);
		END;`,
		// Trigger for DELETE
		`CREATE TRIGGER IF NOT EXISTS messages_fts_ad AFTER DELETE ON messages BEGIN
			INSERT INTO messages_fts(messages_fts, rowid, session_id, role, content)
			VALUES ('delete', old.id, old.session_id, old.role, old.content);
		END;`,
		// Trigger for UPDATE
		`CREATE TRIGGER IF NOT EXISTS messages_fts_au AFTER UPDATE ON messages BEGIN
			INSERT INTO messages_fts(messages_fts, rowid, session_id, role, content)
			VALUES ('delete', old.id, old.session_id, old.role, old.content);
			INSERT INTO messages_fts(rowid, session_id, role, content)
			VALUES (new.id, new.session_id, new.role, new.content);
		END;`,
	}

	for _, trigger := range fts5Triggers {
		if _, err := db.conn.Exec(trigger); err != nil {
			// Ignore errors if trigger already exists
			if !strings.Contains(err.Error(), "already exists") {
				log.Printf("Warning: failed to create FTS5 trigger: %v", err)
			}
		}
	}

	// Populate FTS5 index from existing messages if needed
	if err := db.migrateFTS5(); err != nil {
		log.Printf("Warning: failed to populate FTS5 index: %v", err)
	}

	return nil
}

// migrateMessagesTableForThinking recreates the messages table to allow 'thinking' role
func (db *DB) migrateMessagesTableForThinking() error {
	// Check if migration is needed by trying to insert a thinking message
	_, err := db.conn.Exec("INSERT INTO messages (session_id, role, content, created_at) VALUES ('__migration_test__', 'thinking', 'test', datetime('now'))")
	if err == nil {
		// Migration not needed, clean up test row
		db.conn.Exec("DELETE FROM messages WHERE session_id = '__migration_test__'")
		return nil
	}

	// If the error contains "CHECK constraint", we need to migrate
	// Use Contains for robustness across different SQLite driver versions
	errStr := strings.ToLower(err.Error())
	if strings.Contains(errStr, "check") && strings.Contains(errStr, "constraint") {
		log.Println("Migrating messages table to allow 'thinking' role...")

		tx, err := db.conn.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer tx.Rollback()

		// Create new table with updated constraint
		_, err = tx.Exec(`
			CREATE TABLE IF NOT EXISTS messages_new (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				session_id TEXT NOT NULL,
				role TEXT NOT NULL CHECK(role IN ('user', 'assistant', 'thinking')),
				content TEXT NOT NULL,
				created_at DATETIME NOT NULL,
				FOREIGN KEY (session_id) REFERENCES sessions(session_id) ON DELETE CASCADE
			)
		`)
		if err != nil {
			return fmt.Errorf("failed to create new messages table: %w", err)
		}

		// Copy data
		_, err = tx.Exec(`
			INSERT INTO messages_new (id, session_id, role, content, created_at)
			SELECT id, session_id, role, content, created_at FROM messages
		`)
		if err != nil {
			return fmt.Errorf("failed to copy messages data: %w", err)
		}

		// Drop old table
		_, err = tx.Exec("DROP TABLE messages")
		if err != nil {
			return fmt.Errorf("failed to drop old messages table: %w", err)
		}

		// Rename new table
		_, err = tx.Exec("ALTER TABLE messages_new RENAME TO messages")
		if err != nil {
			return fmt.Errorf("failed to rename messages table: %w", err)
		}

		// Recreate indexes
		_, err = tx.Exec("CREATE INDEX IF NOT EXISTS idx_messages_session_id ON messages(session_id)")
		if err != nil {
			return fmt.Errorf("failed to create session_id index: %w", err)
		}

		_, err = tx.Exec("CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at)")
		if err != nil {
			return fmt.Errorf("failed to create created_at index: %w", err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		log.Println("Messages table migration completed successfully")
	}

	return nil
}

// migrateFTS5 populates the FTS5 index from existing messages if needed
func (db *DB) migrateFTS5() error {
	// Check if FTS table is empty but messages exist
	var ftsCount, msgCount int
	if err := db.conn.QueryRow("SELECT COUNT(*) FROM messages_fts").Scan(&ftsCount); err != nil {
		return fmt.Errorf("failed to count FTS entries: %w", err)
	}
	if err := db.conn.QueryRow("SELECT COUNT(*) FROM messages").Scan(&msgCount); err != nil {
		return fmt.Errorf("failed to count messages: %w", err)
	}

	if ftsCount == 0 && msgCount > 0 {
		log.Printf("Populating FTS5 index from %d existing messages...", msgCount)
		_, err := db.conn.Exec(`
			INSERT INTO messages_fts(rowid, session_id, role, content)
			SELECT id, session_id, role, content FROM messages
		`)
		if err != nil {
			return fmt.Errorf("failed to populate FTS5: %w", err)
		}
		log.Printf("FTS5 index populated with %d messages", msgCount)
	}

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
