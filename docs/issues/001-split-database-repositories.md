# Split database.go into domain repositories

**Priority:** P1 (High)
**Type:** Refactoring
**Component:** Backend
**Estimated Effort:** Medium

## Summary

Split the monolithic `models/database.go` (1441 lines) into domain-specific repositories following the repository pattern.

## Current State

All database operations are in a single file mixing concerns:
- Session CRUD
- Message CRUD
- Memory entries
- Tool calls
- SSH machines
- Settings
- Full-text search

## Proposed Structure

```
backend/
├── repositories/
│   ├── repository.go        # Common interface & base
│   ├── session_repo.go      # Session CRUD (~150 lines)
│   ├── message_repo.go      # Message CRUD (~100 lines)
│   ├── memory_repo.go       # Memory entries (~150 lines)
│   ├── machine_repo.go      # SSH machines (~200 lines)
│   ├── tool_call_repo.go    # Tool calls (~100 lines)
│   ├── settings_repo.go     # Settings (~80 lines)
│   └── search_repo.go       # FTS search (~80 lines)
├── models/
│   ├── session.go           # Session struct
│   ├── message.go           # Message struct
│   ├── memory.go            # MemoryEntry struct
│   ├── machine.go           # Machine struct
│   ├── tool_call.go         # ToolCall struct
│   └── search.go            # SearchResult struct
```

## Example Repository Interface

```go
// repositories/repository.go
package repositories

type SessionRepository interface {
    Create(sessionID, model string) (*models.Session, error)
    Get(sessionID string) (*models.Session, error)
    List() ([]*models.Session, error)
    UpdateTitle(sessionID, title string) error
    UpdateModel(sessionID, model string) error
    UpdateActivity(sessionID string) error
    Delete(sessionID string) error
}

type MessageRepository interface {
    Save(sessionID, role, content string) (*models.Message, error)
    GetBySession(sessionID string) ([]*models.Message, error)
    Search(query string, limit, offset int) ([]*models.SearchResult, int, error)
}
```

## Tasks

- [ ] Create `repositories/` package
- [ ] Define repository interfaces (SessionRepository, MessageRepository, etc.)
- [ ] Extract session operations to `session_repo.go`
- [ ] Extract message operations to `message_repo.go`
- [ ] Extract memory operations to `memory_repo.go`
- [ ] Extract machine operations to `machine_repo.go`
- [ ] Extract tool call operations to `tool_call_repo.go`
- [ ] Extract settings operations to `settings_repo.go`
- [ ] Extract search operations to `search_repo.go`
- [ ] Split model structs into separate files
- [ ] Update handlers to use repository interfaces
- [ ] Add unit tests for repositories

## Acceptance Criteria

- [ ] Each repository file is under 200 lines
- [ ] All existing tests pass
- [ ] Repository interfaces allow for mocking in tests
- [ ] `models/database.go` only contains DB connection and initialization

## References

- `ARCHITECTURE_REVIEW.md` section "1. Split `models/database.go`"
- Current file: `backend/models/database.go`

## Labels

```
priority: P1
type: refactoring
component: backend
```
