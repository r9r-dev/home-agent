# Extract prompt builders from chat.go

**Priority:** P1 (High)
**Type:** Refactoring
**Component:** Backend
**Estimated Effort:** Medium

## Summary

Refactor `handlers/chat.go` (696 lines) by extracting prompt building, attachment handling, and context injection into focused service components.

## Current State

Single handler mixing multiple responsibilities:
- Prompt building
- Attachment handling (images, files)
- SSH context injection
- Memory injection
- Response processing

## Proposed Structure

```
backend/
├── handlers/
│   └── chat.go              # Slim handler (~150 lines)
├── services/
│   ├── prompt/
│   │   ├── builder.go       # PromptBuilder interface
│   │   ├── attachment.go    # AttachmentProcessor
│   │   ├── memory.go        # MemoryInjector
│   │   └── ssh_context.go   # SSHContextBuilder
│   └── response/
│       └── processor.go     # ResponseProcessor
```

## Example PromptBuilder

```go
// services/prompt/builder.go
package prompt

type Builder struct {
    attachmentProcessor AttachmentProcessor
    memoryInjector      MemoryInjector
    sshContextBuilder   SSHContextBuilder
}

func (b *Builder) Build(request MessageRequest) (*BuiltPrompt, error) {
    prompt := &BuiltPrompt{}

    // Each step is now testable independently
    if err := b.attachmentProcessor.Process(request, prompt); err != nil {
        return nil, err
    }

    if err := b.memoryInjector.Inject(prompt); err != nil {
        return nil, err
    }

    if request.MachineID != "" {
        if err := b.sshContextBuilder.AddContext(request.MachineID, prompt); err != nil {
            return nil, err
        }
    }

    return prompt, nil
}
```

## Tasks

- [ ] Create `services/prompt/` package
- [ ] Extract `AttachmentProcessor` for file/image handling
- [ ] Extract `MemoryInjector` for memory context
- [ ] Extract `SSHContextBuilder` for SSH machine context
- [ ] Create `PromptBuilder` that orchestrates all processors
- [ ] Extract `ResponseProcessor` for Claude response handling
- [ ] Update `ChatHandler` to use new services
- [ ] Add unit tests for each component
- [ ] Ensure all path mapping logic is centralized

## Acceptance Criteria

- [ ] `chat.go` is under 200 lines
- [ ] Each service component is independently testable
- [ ] All existing functionality preserved
- [ ] Clear interfaces between components

## References

- `ARCHITECTURE_REVIEW.md` section "3. Refactor `handlers/chat.go`"
- Current file: `backend/handlers/chat.go`

## Labels

```
priority: P1
type: refactoring
component: backend
```
