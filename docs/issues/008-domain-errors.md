# Add domain error types

**Priority:** P2 (Medium)
**Type:** Enhancement
**Component:** Backend
**Estimated Effort:** Low

## Summary

Replace raw `fmt.Errorf` calls with typed domain errors that include HTTP status codes and error codes.

## Current State

```go
// Scattered error handling with no structure
return fmt.Errorf("session not found: %s", sessionID)
return fmt.Errorf("failed to create session: %w", err)
return c.Status(404).JSON(fiber.Map{"error": "Session not found"})
```

## Proposed Solution

```go
// pkg/errors/errors.go
package errors

import "fmt"

// Predefined domain errors
var (
    ErrSessionNotFound = &DomainError{Code: "SESSION_NOT_FOUND", Status: 404}
    ErrMachineNotFound = &DomainError{Code: "MACHINE_NOT_FOUND", Status: 404}
    ErrMemoryNotFound  = &DomainError{Code: "MEMORY_NOT_FOUND", Status: 404}
    ErrInvalidInput    = &DomainError{Code: "INVALID_INPUT", Status: 400}
    ErrProxyConnection = &DomainError{Code: "PROXY_CONNECTION_FAILED", Status: 503}
    ErrUploadFailed    = &DomainError{Code: "UPLOAD_FAILED", Status: 500}
    ErrFileTooLarge    = &DomainError{Code: "FILE_TOO_LARGE", Status: 413}
    ErrInvalidFileType = &DomainError{Code: "INVALID_FILE_TYPE", Status: 415}
)

type DomainError struct {
    Code    string `json:"code"`
    Status  int    `json:"-"`
    Message string `json:"message,omitempty"`
    Cause   error  `json:"-"`
}

func (e *DomainError) Error() string {
    if e.Message != "" {
        return e.Message
    }
    return e.Code
}

func (e *DomainError) WithMessage(msg string) *DomainError {
    return &DomainError{
        Code:    e.Code,
        Status:  e.Status,
        Message: msg,
        Cause:   e.Cause,
    }
}

func (e *DomainError) WithCause(err error) *DomainError {
    return &DomainError{
        Code:    e.Code,
        Status:  e.Status,
        Message: e.Message,
        Cause:   err,
    }
}

func (e *DomainError) Unwrap() error {
    return e.Cause
}
```

## Error Handling Middleware

```go
// middleware/error_handler.go
func ErrorHandler(c *fiber.Ctx, err error) error {
    var domainErr *errors.DomainError
    if errors.As(err, &domainErr) {
        return c.Status(domainErr.Status).JSON(fiber.Map{
            "error": domainErr.Code,
            "message": domainErr.Message,
        })
    }

    // Default to 500 for unknown errors
    return c.Status(500).JSON(fiber.Map{
        "error": "INTERNAL_ERROR",
        "message": "An unexpected error occurred",
    })
}
```

## Usage

```go
// In repository
func (r *SessionRepo) Get(sessionID string) (*models.Session, error) {
    // ...
    if err == sql.ErrNoRows {
        return nil, errors.ErrSessionNotFound.WithMessage(
            fmt.Sprintf("session %s not found", sessionID),
        )
    }
    return nil, errors.ErrInternalError.WithCause(err)
}
```

## Tasks

- [ ] Create `pkg/errors/` package
- [ ] Define `DomainError` type with methods
- [ ] Define error constants for all domains
- [ ] Add error handling middleware to Fiber
- [ ] Update handlers to return domain errors
- [ ] Update repositories to return domain errors
- [ ] Add error code to API responses
- [ ] Document error codes in API docs

## Acceptance Criteria

- [ ] All errors have consistent structure
- [ ] HTTP status derived from error type
- [ ] Error codes are documented
- [ ] Stack traces preserved for debugging (via Cause)

## References

- `ARCHITECTURE_REVIEW.md` section "6. Add Domain Errors"

## Labels

```
priority: P2
type: enhancement
component: backend
```
