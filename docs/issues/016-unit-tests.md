# Add unit tests for core components

**Priority:** P2 (Medium)
**Type:** Enhancement
**Component:** Backend
**Estimated Effort:** High

## Summary

Establish unit testing practices with comprehensive tests for repositories, services, and handlers.

## Focus Areas

1. **Repository layer** - Use SQLite in-memory database
2. **Prompt building logic** - Test context injection
3. **Response processing** - Test stream handling
4. **Service layer** - Test business logic

## Test Infrastructure

```go
// testutil/database.go
package testutil

import (
    "database/sql"
    "testing"

    "home-agent/internal/database"
)

func SetupTestDB(t *testing.T) *sql.DB {
    t.Helper()

    db, err := sql.Open("sqlite", ":memory:")
    if err != nil {
        t.Fatalf("failed to open test database: %v", err)
    }

    if err := database.RunMigrations(db); err != nil {
        t.Fatalf("failed to run migrations: %v", err)
    }

    t.Cleanup(func() {
        db.Close()
    })

    return db
}
```

## Example Repository Tests

```go
// repositories/session_repo_test.go
package repositories

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "home-agent/testutil"
)

func TestSessionRepository_Create(t *testing.T) {
    db := testutil.SetupTestDB(t)
    repo := NewSessionRepository(db)

    session, err := repo.Create("test-id", "haiku")

    require.NoError(t, err)
    assert.Equal(t, "test-id", session.SessionID)
    assert.Equal(t, "haiku", session.Model)
    assert.NotZero(t, session.CreatedAt)
}

func TestSessionRepository_Get_NotFound(t *testing.T) {
    db := testutil.SetupTestDB(t)
    repo := NewSessionRepository(db)

    session, err := repo.Get("nonexistent")

    assert.NoError(t, err)
    assert.Nil(t, session)
}

func TestSessionRepository_Delete(t *testing.T) {
    db := testutil.SetupTestDB(t)
    repo := NewSessionRepository(db)

    // Create session first
    _, err := repo.Create("test-id", "haiku")
    require.NoError(t, err)

    // Delete it
    err = repo.Delete("test-id")
    require.NoError(t, err)

    // Verify it's gone
    session, err := repo.Get("test-id")
    assert.NoError(t, err)
    assert.Nil(t, session)
}
```

## Example Handler Tests

```go
// handlers/memory_handler_test.go
package handlers

import (
    "net/http/httptest"
    "testing"

    "github.com/gofiber/fiber/v2"
    "github.com/stretchr/testify/assert"
)

type mockMemoryRepo struct {
    entries []*models.MemoryEntry
}

func (m *mockMemoryRepo) List() ([]*models.MemoryEntry, error) {
    return m.entries, nil
}

func TestMemoryHandler_List(t *testing.T) {
    mock := &mockMemoryRepo{
        entries: []*models.MemoryEntry{
            {ID: "1", Title: "Test", Content: "Content"},
        },
    }
    handler := NewMemoryHandler(mock)

    app := fiber.New()
    app.Get("/api/memory", handler.List)

    req := httptest.NewRequest("GET", "/api/memory", nil)
    resp, err := app.Test(req)

    assert.NoError(t, err)
    assert.Equal(t, 200, resp.StatusCode)
}
```

## Example Prompt Builder Tests

```go
// services/prompt/builder_test.go
package prompt

import (
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestPromptBuilder_WithMemory(t *testing.T) {
    builder := NewBuilder()
    memory := []*models.MemoryEntry{
        {Title: "User Info", Content: "Name: John"},
    }

    prompt, err := builder.WithMemory(memory).Build("Hello")

    assert.NoError(t, err)
    assert.Contains(t, prompt, "<user_memory>")
    assert.Contains(t, prompt, "Name: John")
}

func TestPromptBuilder_WithAttachments(t *testing.T) {
    builder := NewBuilder()
    attachments := []types.Attachment{
        {Type: "file", Filename: "test.txt", Path: "/tmp/test.txt"},
    }

    prompt, err := builder.WithAttachments(attachments).Build("Analyze this")

    assert.NoError(t, err)
    assert.Contains(t, prompt, "test.txt")
}
```

## Coverage Targets

| Component | Target Coverage |
|-----------|----------------|
| Repositories | 80% |
| Services | 70% |
| Handlers | 60% |
| Overall | 65% |

## Tasks

- [ ] Set up test infrastructure
- [ ] Create in-memory SQLite helper
- [ ] Add testify dependency for assertions
- [ ] Write repository tests (session, message, memory, machine)
- [ ] Write service tests (prompt builder, session manager)
- [ ] Write handler tests with mocks
- [ ] Add CI test step to GitHub Actions
- [ ] Configure coverage reporting
- [ ] Add test documentation

## Acceptance Criteria

- [ ] All repositories have tests
- [ ] Key services have tests
- [ ] Handlers have integration tests
- [ ] CI runs tests on every PR
- [ ] Coverage report generated
- [ ] Minimum 60% overall coverage

## References

- `ARCHITECTURE_REVIEW.md` section "3. Add Unit Tests"
- testify: https://github.com/stretchr/testify

## Labels

```
priority: P2
type: enhancement
component: backend
```
