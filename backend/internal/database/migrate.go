package database

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// LatestVersion is the current migration version
// Update this when adding new migrations
const LatestVersion = 9

// Migrate runs database migrations
func (db *DB) Migrate() error {
	log.Println("Starting database migrations...")

	// Check if this is an existing database without schema_migrations
	isLegacy, err := db.isLegacyDatabase()
	if err != nil {
		return fmt.Errorf("failed to check legacy database: %w", err)
	}

	// Run migrations using a separate connection to avoid connection closure issues
	// The sqlite driver closes the connection when the migrator is closed
	if err := db.runMigrations(isLegacy); err != nil {
		return err
	}

	// Populate FTS5 index if needed (using main connection)
	if err := db.migrateFTS5(); err != nil {
		log.Printf("Warning: failed to populate FTS5 index: %v", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// runMigrations executes the actual migration using a separate connection
func (db *DB) runMigrations(isLegacy bool) error {
	// Open a separate connection for migrations
	dsn := db.path + "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"
	migrationConn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return fmt.Errorf("failed to open migration connection: %w", err)
	}
	defer migrationConn.Close()

	// Create migration source from embedded files
	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create migration source: %w", err)
	}

	// Create migration driver with separate connection
	driver, err := sqlite.WithInstance(migrationConn, &sqlite.Config{
		MigrationsTable: "schema_migrations",
		NoTxWrap:        false,
	})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	// Create migrator
	m, err := migrate.NewWithInstance("iofs", source, "sqlite", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer m.Close()

	// For legacy databases, force the version to latest since schema is already complete
	if isLegacy {
		log.Printf("Legacy database detected, setting version to %d", LatestVersion)
		if err := m.Force(LatestVersion); err != nil {
			return fmt.Errorf("failed to force migration version: %w", err)
		}
		log.Println("Legacy database migrated successfully")
		return nil
	}

	// Run migrations
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	version, dirty, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return fmt.Errorf("failed to get migration version: %w", err)
	}

	if dirty {
		log.Printf("Warning: database is in dirty state at version %d", version)
	} else if errors.Is(err, migrate.ErrNilVersion) {
		log.Println("Database initialized with no migrations")
	} else {
		log.Printf("Database at migration version %d", version)
	}

	return nil
}

// isLegacyDatabase checks if this is a pre-migration database
// (has tables but no schema_migrations table)
func (db *DB) isLegacyDatabase() (bool, error) {
	// Check if schema_migrations table exists
	var schemaMigrationsExists bool
	err := db.conn.QueryRow(`
		SELECT COUNT(*) > 0 FROM sqlite_master
		WHERE type='table' AND name='schema_migrations'
	`).Scan(&schemaMigrationsExists)
	if err != nil {
		return false, fmt.Errorf("failed to check schema_migrations: %w", err)
	}

	if schemaMigrationsExists {
		return false, nil
	}

	// Check if sessions table exists (indicator of existing data)
	var sessionsExists bool
	err = db.conn.QueryRow(`
		SELECT COUNT(*) > 0 FROM sqlite_master
		WHERE type='table' AND name='sessions'
	`).Scan(&sessionsExists)
	if err != nil {
		return false, fmt.Errorf("failed to check sessions table: %w", err)
	}

	return sessionsExists, nil
}

// migrateFTS5 populates the FTS5 index from existing messages if needed
func (db *DB) migrateFTS5() error {
	// Check if FTS table exists
	var ftsExists bool
	err := db.conn.QueryRow(`
		SELECT COUNT(*) > 0 FROM sqlite_master
		WHERE type='table' AND name='messages_fts'
	`).Scan(&ftsExists)
	if err != nil {
		return fmt.Errorf("failed to check FTS table: %w", err)
	}

	if !ftsExists {
		return nil
	}

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

// Version returns the current migration version
func (db *DB) Version() (uint, bool, error) {
	var version int
	var dirty bool

	err := db.conn.QueryRow(`
		SELECT version, dirty FROM schema_migrations LIMIT 1
	`).Scan(&version, &dirty)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, false, migrate.ErrNilVersion
		}
		return 0, false, err
	}

	return uint(version), dirty, nil
}

// MigrateDown rolls back n migrations (0 means all)
func (db *DB) MigrateDown(steps int) error {
	// Use separate connection to avoid closure issues
	dsn := db.path + "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"
	migrationConn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return fmt.Errorf("failed to open migration connection: %w", err)
	}
	defer migrationConn.Close()

	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create migration source: %w", err)
	}

	driver, err := sqlite.WithInstance(migrationConn, &sqlite.Config{
		MigrationsTable: "schema_migrations",
	})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "sqlite", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer m.Close()

	if steps == 0 {
		return m.Down()
	}
	return m.Steps(-steps)
}

// MigrateTo migrates to a specific version
func (db *DB) MigrateTo(version uint) error {
	// Use separate connection to avoid closure issues
	dsn := db.path + "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"
	migrationConn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return fmt.Errorf("failed to open migration connection: %w", err)
	}
	defer migrationConn.Close()

	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create migration source: %w", err)
	}

	driver, err := sqlite.WithInstance(migrationConn, &sqlite.Config{
		MigrationsTable: "schema_migrations",
	})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "sqlite", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer m.Close()

	return m.Migrate(version)
}

// ForceVersion forces the migration version without running migrations
// Use with caution - this can leave the database in an inconsistent state
func (db *DB) ForceVersion(version int) error {
	// Use separate connection to avoid closure issues
	dsn := db.path + "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"
	migrationConn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return fmt.Errorf("failed to open migration connection: %w", err)
	}
	defer migrationConn.Close()

	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create migration source: %w", err)
	}

	driver, err := sqlite.WithInstance(migrationConn, &sqlite.Config{
		MigrationsTable: "schema_migrations",
	})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "sqlite", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer m.Close()

	return m.Force(version)
}

// tableExists checks if a table exists in the database
func (db *DB) tableExists(tableName string) (bool, error) {
	var exists bool
	err := db.conn.QueryRow(`
		SELECT COUNT(*) > 0 FROM sqlite_master
		WHERE type='table' AND name=?
	`, tableName).Scan(&exists)
	return exists, err
}

// columnExists checks if a column exists in a table
func (db *DB) columnExists(tableName, columnName string) (bool, error) {
	rows, err := db.conn.Query(fmt.Sprintf("PRAGMA table_info(%s)", tableName))
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dfltValue sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
			return false, err
		}
		if name == columnName {
			return true, nil
		}
	}
	return false, rows.Err()
}
