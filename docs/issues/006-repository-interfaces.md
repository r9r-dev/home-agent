# Add repository interfaces for testability

**Priority:** P2 (Medium)
**Type:** Refactoring
**Component:** Backend
**Estimated Effort:** Medium

## Summary

Replace direct `*models.DB` dependencies with interfaces to enable unit testing with mocks.

## Current State

```go
// Current (hard to test)
type MemoryHandler struct {
    db *models.DB
}

func (h *MemoryHandler) List(c *fiber.Ctx) error {
    entries, err := h.db.ListMemoryEntries()  // Direct DB call
    // ...
}
```

## Proposed Solution

```go
// Better (testable)
type MemoryHandler struct {
    repo repositories.MemoryRepository
}

func (h *MemoryHandler) List(c *fiber.Ctx) error {
    entries, err := h.repo.List()  // Interface call
    // ...
}

// Mock for tests
type mockMemoryRepo struct{}

func (m *mockMemoryRepo) List() ([]*models.MemoryEntry, error) {
    return []*models.MemoryEntry{
        {ID: "1", Title: "Test", Content: "Content"},
    }, nil
}

func (m *mockMemoryRepo) Create(id, title, content string) (*models.MemoryEntry, error) {
    return &models.MemoryEntry{ID: id, Title: title, Content: content}, nil
}
```

## Example Test

```go
func TestMemoryHandler_List(t *testing.T) {
    mock := &mockMemoryRepo{}
    handler := NewMemoryHandler(mock)

    app := fiber.New()
    app.Get("/api/memory", handler.List)

    req := httptest.NewRequest("GET", "/api/memory", nil)
    resp, _ := app.Test(req)

    assert.Equal(t, 200, resp.StatusCode)
}
```

## Tasks

- [ ] Define repository interfaces for all domains
- [ ] Update all handlers to accept interfaces
- [ ] Create constructor functions for dependency injection
- [ ] Create mock implementations for tests
- [ ] Write unit tests using mocks
- [ ] Document testing patterns

## Acceptance Criteria

- [ ] All handlers use interfaces, not concrete types
- [ ] Mock implementations exist for all repositories
- [ ] At least 50% test coverage on handlers
- [ ] No direct DB access from handlers

## References

- `ARCHITECTURE_REVIEW.md` section "8. Add Interfaces for Testability"
- Depends on: Issue #001 (Split database.go)

## Labels

```
priority: P2
type: refactoring
component: backend
```
