# Standardize logging across backend

**Priority:** P3 (Low)
**Type:** Enhancement
**Component:** Backend
**Estimated Effort:** Medium

## Summary

Replace scattered `log.Printf` calls with structured logging for better observability and consistency.

## Current State

Mixed logging approaches throughout the codebase:

```go
log.Printf("Created new session: %s (ID: %d, model: %s)", sessionID, id, model)
log.Printf("Saved %s message for session %s (ID: %d)", role, sessionID, id)
log.Println("Database initialized successfully")
```

Also uses `logService` for user-visible logs with different format.

## Proposed Solution

Use Go 1.21+ `slog` (standard library) for structured logging:

```go
// pkg/logging/logger.go
package logging

import (
    "log/slog"
    "os"
)

var Logger *slog.Logger

func Init(level string) {
    opts := &slog.HandlerOptions{
        Level: parseLevel(level),
    }

    handler := slog.NewJSONHandler(os.Stdout, opts)
    Logger = slog.New(handler)
    slog.SetDefault(Logger)
}

func parseLevel(level string) slog.Level {
    switch level {
    case "debug":
        return slog.LevelDebug
    case "warn":
        return slog.LevelWarn
    case "error":
        return slog.LevelError
    default:
        return slog.LevelInfo
    }
}
```

## Usage Examples

```go
// Before
log.Printf("Created new session: %s (ID: %d, model: %s)", sessionID, id, model)

// After
slog.Info("session created",
    slog.String("session_id", sessionID),
    slog.Int("id", id),
    slog.String("model", model),
)

// Before
log.Printf("Error processing message: %v", err)

// After
slog.Error("message processing failed",
    slog.String("session_id", sessionID),
    slog.Any("error", err),
)
```

## Output Format

```json
{
  "time": "2024-01-15T10:30:00Z",
  "level": "INFO",
  "msg": "session created",
  "session_id": "abc-123",
  "id": 42,
  "model": "haiku"
}
```

## Request Context

Add request ID to all logs:

```go
// middleware/request_id.go
func RequestIDMiddleware() fiber.Handler {
    return func(c *fiber.Ctx) error {
        requestID := uuid.New().String()
        c.Locals("request_id", requestID)

        // Add to logger context
        logger := slog.With(slog.String("request_id", requestID))
        c.Locals("logger", logger)

        return c.Next()
    }
}

// Usage in handler
func (h *Handler) Get(c *fiber.Ctx) error {
    logger := c.Locals("logger").(*slog.Logger)
    logger.Info("fetching session", slog.String("session_id", id))
    // ...
}
```

## Tasks

- [ ] Add `pkg/logging/` package
- [ ] Create logger initialization function
- [ ] Add log level configuration
- [ ] Add request ID middleware
- [ ] Update all `log.Printf` calls to `slog`
- [ ] Add context-aware logging to handlers
- [ ] Configure JSON output for production
- [ ] Configure text output for development
- [ ] Document logging standards

## Acceptance Criteria

- [ ] All logs use structured format
- [ ] Request ID in all request logs
- [ ] Log levels configurable via environment
- [ ] JSON format for production
- [ ] Human-readable format for development
- [ ] Consistent field names across codebase

## References

- `ARCHITECTURE_REVIEW.md` section "2. Standardize Logging"
- Go slog: https://pkg.go.dev/log/slog

## Labels

```
priority: P3
type: enhancement
component: backend
```
