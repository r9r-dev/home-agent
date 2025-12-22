package database

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMigrations_NewDatabase(t *testing.T) {
	// Create temporary directory for test database
	tmpDir, err := os.MkdirTemp("", "home-agent-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")

	// Create new database connection
	db, err := New(Config{Path: dbPath})
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.Migrate(); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Verify version
	version, dirty, err := db.Version()
	if err != nil {
		t.Fatalf("Failed to get version: %v", err)
	}

	if dirty {
		t.Error("Database should not be in dirty state")
	}

	if version != LatestVersion {
		t.Errorf("Expected version %d, got %d", LatestVersion, version)
	}

	// Verify tables exist
	tables := []string{"sessions", "messages", "settings", "memory", "tool_calls", "machines", "messages_fts"}
	for _, table := range tables {
		exists, err := db.tableExists(table)
		if err != nil {
			t.Errorf("Failed to check table %s: %v", table, err)
		}
		if !exists {
			t.Errorf("Table %s should exist", table)
		}
	}

	// Verify sessions table has all columns
	columns := []string{"id", "session_id", "title", "claude_session_id", "model", "created_at", "last_activity"}
	for _, col := range columns {
		exists, err := db.columnExists("sessions", col)
		if err != nil {
			t.Errorf("Failed to check column %s: %v", col, err)
		}
		if !exists {
			t.Errorf("Column sessions.%s should exist", col)
		}
	}

	// Verify messages table allows 'thinking' role
	_, err = db.conn.Exec(`
		INSERT INTO messages (session_id, role, content, created_at)
		VALUES ('test-session', 'thinking', 'test content', datetime('now'))
	`)
	if err != nil {
		t.Errorf("Should be able to insert thinking role: %v", err)
	}
}

func TestMigrations_LegacyDatabase(t *testing.T) {
	// Create temporary directory for test database
	tmpDir, err := os.MkdirTemp("", "home-agent-legacy-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "legacy.db")

	// Create a "legacy" database with existing tables but no schema_migrations
	db, err := New(Config{Path: dbPath})
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	// Create tables manually (simulating legacy database)
	_, err = db.conn.Exec(`
		CREATE TABLE sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT UNIQUE NOT NULL,
			claude_session_id TEXT DEFAULT '',
			title TEXT DEFAULT '',
			model TEXT DEFAULT 'haiku',
			created_at DATETIME NOT NULL,
			last_activity DATETIME NOT NULL
		);
		CREATE TABLE messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL,
			role TEXT NOT NULL CHECK(role IN ('user', 'assistant', 'thinking')),
			content TEXT NOT NULL,
			created_at DATETIME NOT NULL
		);
		CREATE TABLE settings (key TEXT PRIMARY KEY, value TEXT NOT NULL DEFAULT '');
		CREATE TABLE memory (id TEXT PRIMARY KEY, title TEXT NOT NULL, content TEXT NOT NULL, enabled INTEGER DEFAULT 1);
		CREATE TABLE tool_calls (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL,
			tool_use_id TEXT UNIQUE NOT NULL,
			tool_name TEXT NOT NULL,
			input TEXT NOT NULL DEFAULT '{}',
			output TEXT DEFAULT '',
			status TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			completed_at DATETIME
		);
		CREATE TABLE machines (id TEXT PRIMARY KEY, name TEXT NOT NULL, host TEXT NOT NULL, port INTEGER DEFAULT 22, username TEXT NOT NULL, auth_type TEXT NOT NULL, auth_value TEXT NOT NULL);
		CREATE VIRTUAL TABLE IF NOT EXISTS messages_fts USING fts5(session_id UNINDEXED, role UNINDEXED, content);
	`)
	if err != nil {
		t.Fatalf("Failed to create legacy tables: %v", err)
	}

	// Insert some test data
	_, err = db.conn.Exec(`
		INSERT INTO sessions (session_id, created_at, last_activity) VALUES ('legacy-session', datetime('now'), datetime('now'));
		INSERT INTO messages (session_id, role, content, created_at) VALUES ('legacy-session', 'user', 'Hello', datetime('now'));
	`)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Close and reopen to simulate fresh start
	db.Close()

	// Reopen database
	db, err = New(Config{Path: dbPath})
	if err != nil {
		t.Fatalf("Failed to reopen database: %v", err)
	}
	defer db.Close()

	// Run migrations - should detect legacy database and force version
	if err := db.Migrate(); err != nil {
		t.Fatalf("Failed to migrate legacy database: %v", err)
	}

	// Verify version is set to latest
	version, dirty, err := db.Version()
	if err != nil {
		t.Fatalf("Failed to get version: %v", err)
	}

	if dirty {
		t.Error("Database should not be in dirty state")
	}

	if version != LatestVersion {
		t.Errorf("Expected version %d, got %d", LatestVersion, version)
	}

	// Verify data is preserved
	var count int
	err = db.conn.QueryRow("SELECT COUNT(*) FROM sessions WHERE session_id = 'legacy-session'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query sessions: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 legacy session, got %d", count)
	}
}

func TestMigrations_Idempotent(t *testing.T) {
	// Create temporary directory for test database
	tmpDir, err := os.MkdirTemp("", "home-agent-idempotent-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "idempotent.db")

	// Create and migrate database
	db, err := New(Config{Path: dbPath})
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	if err := db.Migrate(); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Insert test data
	_, err = db.conn.Exec(`
		INSERT INTO sessions (session_id, created_at, last_activity) VALUES ('test-session', datetime('now'), datetime('now'));
		INSERT INTO messages (session_id, role, content, created_at) VALUES ('test-session', 'user', 'Hello', datetime('now'));
	`)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Close and reopen
	db.Close()

	db, err = New(Config{Path: dbPath})
	if err != nil {
		t.Fatalf("Failed to reopen database: %v", err)
	}
	defer db.Close()

	// Run migrations again - should be idempotent
	if err := db.Migrate(); err != nil {
		t.Fatalf("Second migration run should succeed: %v", err)
	}

	// Verify data is still there
	var count int
	err = db.conn.QueryRow("SELECT COUNT(*) FROM sessions").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query sessions: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 session, got %d", count)
	}
}
