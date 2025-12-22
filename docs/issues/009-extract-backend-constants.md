# Extract backend constants and magic numbers

**Priority:** P2 (Medium)
**Type:** Refactoring
**Component:** Backend
**Estimated Effort:** Low
**Good First Issue:** Yes

## Summary

Centralize magic numbers and strings scattered throughout the backend codebase into dedicated constants.

## Examples of Magic Values Found

```go
// Size limits
if len(content) > 2000 { ... }                    // Max custom instructions
if file.Size > 10*1024*1024 { ... }               // Max file size (10MB)
if len(content) > 100*1024 { ... }                // Max file content (100KB)

// Buffer sizes
responseChan := make(chan services.ClaudeResponse, 100)

// Model names
model = "haiku"
model = "sonnet"
model = "opus"

// Timeouts
time.Sleep(10 * time.Minute)
```

## Proposed Solution

```go
// constants/limits.go
package constants

import "time"

const (
    // Size limits
    MaxCustomInstructionsLength = 2000
    MaxFileSizeBytes           = 10 * 1024 * 1024  // 10MB
    MaxFileContentBytes        = 100 * 1024        // 100KB for prompt embedding

    // Buffer sizes
    LogBufferSize       = 100
    WebSocketBufferSize = 100
    ResponseChannelSize = 100

    // Timeouts
    ClaudeTimeout      = 10 * time.Minute
    WebSocketPingInterval = 30 * time.Second
    DatabaseBusyTimeout   = 5000 // milliseconds
)

// constants/models.go
package constants

const (
    ModelHaiku  = "haiku"
    ModelSonnet = "sonnet"
    ModelOpus   = "opus"

    DefaultModel = ModelHaiku
)

// constants/roles.go
package constants

const (
    RoleUser      = "user"
    RoleAssistant = "assistant"
    RoleThinking  = "thinking"
)

// constants/status.go
package constants

const (
    StatusRunning = "running"
    StatusSuccess = "success"
    StatusError   = "error"

    MachineStatusUntested = "untested"
    MachineStatusOnline   = "online"
    MachineStatusOffline  = "offline"
)
```

## Tasks

- [ ] Create `constants/` package
- [ ] Extract size limits to `limits.go`
- [ ] Extract timeout values
- [ ] Extract model names to `models.go`
- [ ] Extract role names to `roles.go`
- [ ] Extract status values to `status.go`
- [ ] Update all usages to reference constants
- [ ] Add documentation comments for each constant

## Acceptance Criteria

- [ ] No magic numbers in business logic
- [ ] All constants are documented
- [ ] Single source of truth for each value
- [ ] Easy to modify values in one place

## References

- `ARCHITECTURE_REVIEW.md` section "5. Extract Magic Numbers"

## Notes

This is a good first issue for new contributors as it:
- Involves straightforward changes
- Improves code quality
- Touches many files but in a simple way
- Has clear acceptance criteria

## Labels

```
priority: P2
type: refactoring
component: backend
good first issue
```
