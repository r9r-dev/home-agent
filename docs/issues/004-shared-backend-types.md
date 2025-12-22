# Create shared backend type definitions

**Priority:** P1 (High)
**Type:** Refactoring
**Component:** Backend
**Estimated Effort:** Low

## Summary

Create a `types/` package to consolidate duplicate type definitions scattered across handlers and services.

## Current State

Types are duplicated across multiple files:
- `Attachment` in `websocket.go`
- `MessageAttachment` in `chat.go`
- `ToolInfo` in `chat.go`
- `ToolCallInfo` in `claude_executor.go`
- `ProxyToolInfo` in `proxy_claude_executor.go`

## Proposed Structure

```go
// pkg/types/attachment.go
package types

type Attachment struct {
    ID       string `json:"id"`
    Filename string `json:"filename"`
    Path     string `json:"path"`
    Type     string `json:"type"` // "image" or "file"
    MimeType string `json:"mime_type,omitempty"`
}

// pkg/types/tool.go
type ToolInfo struct {
    ToolUseID       string                 `json:"tool_use_id"`
    ToolName        string                 `json:"tool_name"`
    Input           map[string]interface{} `json:"input,omitempty"`
    ParentToolUseID string                 `json:"parent_tool_use_id,omitempty"`
}

// pkg/types/message.go
type ClaudeModel string

const (
    ModelHaiku  ClaudeModel = "haiku"
    ModelSonnet ClaudeModel = "sonnet"
    ModelOpus   ClaudeModel = "opus"
)

type MessageRole string

const (
    RoleUser      MessageRole = "user"
    RoleAssistant MessageRole = "assistant"
    RoleThinking  MessageRole = "thinking"
)
```

## Tasks

- [ ] Create `pkg/types/` package
- [ ] Define `Attachment` type
- [ ] Define `ToolInfo` type
- [ ] Define `ClaudeModel` constants
- [ ] Define `MessageRole` constants
- [ ] Define `ToolCallStatus` constants
- [ ] Update all handlers to use shared types
- [ ] Update services to use shared types
- [ ] Remove duplicate definitions

## Acceptance Criteria

- [ ] Single source of truth for each type
- [ ] No duplicate type definitions
- [ ] All imports point to `pkg/types/` package
- [ ] JSON serialization behavior unchanged

## References

- `ARCHITECTURE_REVIEW.md` section "5. Consolidate Type Definitions"

## Labels

```
priority: P1
type: refactoring
component: backend
```
