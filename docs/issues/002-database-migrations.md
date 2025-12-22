# Implement proper database migrations

**Priority:** P1 (High)
**Type:** Architecture
**Component:** Backend
**Estimated Effort:** Medium

## Summary

Replace informal inline `ALTER TABLE` statements with a proper migration system using golang-migrate.

## Current State

Migrations are handled informally with silent failures:

```go
// models/database.go:134-147
alterTableTitle := `ALTER TABLE sessions ADD COLUMN title TEXT DEFAULT '';`
db.conn.Exec(alterTableTitle)  // Errors silently ignored
```

**Problems:**
- No version tracking
- Silent failures hide real errors
- No rollback capability
- Can't verify migration state

## Proposed Solution

Use golang-migrate with embedded SQL files:

```
backend/
├── internal/
│   └── database/
│       ├── db.go              # Connection setup
│       ├── migrate.go         # Migration runner
│       └── migrations/
│           ├── 000001_init_schema.up.sql
│           ├── 000001_init_schema.down.sql
│           ├── 000002_add_thinking_role.up.sql
│           ├── 000002_add_thinking_role.down.sql
│           └── ...
```

## Migration Runner Implementation

```go
// internal/database/migrate.go
package database

import (
    "database/sql"
    "embed"
    "fmt"
    "log"

    "github.com/golang-migrate/migrate/v4"
    "github.com/golang-migrate/migrate/v4/database/sqlite"
    "github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func RunMigrations(db *sql.DB) error {
    driver, err := sqlite.WithInstance(db, &sqlite.Config{})
    if err != nil {
        return fmt.Errorf("create driver: %w", err)
    }

    source, err := iofs.New(migrationsFS, "migrations")
    if err != nil {
        return fmt.Errorf("create source: %w", err)
    }

    m, err := migrate.NewWithInstance("iofs", source, "sqlite", driver)
    if err != nil {
        return fmt.Errorf("create migrator: %w", err)
    }

    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        return fmt.Errorf("run migrations: %w", err)
    }

    version, dirty, _ := m.Version()
    log.Printf("Database at migration version %d (dirty: %v)", version, dirty)

    return nil
}
```

## Example Migration Files

```sql
-- migrations/000001_init_schema.up.sql
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
    role TEXT NOT NULL CHECK(role IN ('user', 'assistant')),
    content TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    FOREIGN KEY (session_id) REFERENCES sessions(session_id) ON DELETE CASCADE
);
```

```sql
-- migrations/000001_init_schema.down.sql
DROP TABLE messages;
DROP TABLE sessions;
```

## SQLite-Specific Considerations

SQLite has limitations:
- No `DROP COLUMN` (until 3.35.0)
- No `ALTER COLUMN`
- No `ADD CONSTRAINT`

Use table recreation pattern for schema changes.

## Tasks

- [ ] Add golang-migrate dependency
- [ ] Create `internal/database/` package
- [ ] Implement migration runner with `//go:embed`
- [ ] Create initial schema migration (000001)
- [ ] Create migration for thinking role (000002)
- [ ] Create migrations for all other schema elements
- [ ] Add `schema_migrations` table
- [ ] Update `main.go` to run migrations on startup
- [ ] Document migration workflow in README
- [ ] Test rollback functionality

## Acceptance Criteria

- [ ] All schema changes are versioned SQL files
- [ ] Migrations run automatically on startup
- [ ] `schema_migrations` table tracks applied versions
- [ ] Rollback is possible with `migrate down`
- [ ] Existing databases are migrated without data loss

## References

- `ARCHITECTURE_REVIEW.md` section "2. Implement Proper Database Migrations"
- golang-migrate: https://github.com/golang-migrate/migrate

## Labels

```
priority: P1
type: architecture
component: backend
```
