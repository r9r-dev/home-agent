# Refactor processMessage in Claude Proxy SDK

**Priority:** P2 (Medium)
**Type:** Refactoring
**Component:** Claude Proxy SDK
**Estimated Effort:** Medium

## Summary

Replace the 230+ line switch statement in `processMessage` with a handler map pattern for better maintainability.

## Current State

Large switch statement in `claude.ts` handling all message types:

```typescript
function processMessage(message: SDKMessage, ctx: ExecutionContext): ProxyResponse | null {
  switch (message.type) {
    case "system":
      // ~20 lines
      break;
    case "assistant":
      // ~30 lines
      break;
    case "stream_event":
      // ~100 lines with nested switch
      break;
    case "tool_progress":
      // ~15 lines
      break;
    case "user":
      // ~60 lines
      break;
    // ...
  }
}
```

## Proposed Solution

```typescript
// src/handlers/messageHandlers.ts
type MessageHandler = (
  message: SDKMessage,
  ctx: ExecutionContext,
  sessionId?: string
) => ProxyResponse | null;

const messageHandlers: Record<string, MessageHandler> = {
  system: handleSystemMessage,
  assistant: handleAssistantMessage,
  stream_event: handleStreamEvent,
  tool_progress: handleToolProgress,
  user: handleUserMessage,
  result: () => null,
};

export function processMessage(
  message: SDKMessage,
  ctx: ExecutionContext,
  sessionId?: string
): ProxyResponse | null {
  const handler = messageHandlers[message.type];
  if (!handler) {
    console.warn(`Unknown message type: ${message.type}`);
    return null;
  }
  return handler(message, ctx, sessionId);
}

// src/handlers/systemHandler.ts
export function handleSystemMessage(
  message: SDKMessage,
  ctx: ExecutionContext
): ProxyResponse | null {
  if (message.subtype === "init") {
    ctx.resetForNewMessage();
    return {
      type: "session_id",
      session_id: message.session_id,
    };
  }
  return null;
}

// src/handlers/streamEventHandler.ts
export function handleStreamEvent(
  message: SDKMessage,
  ctx: ExecutionContext
): ProxyResponse | null {
  const event = message.event;

  const handlers: Record<string, () => ProxyResponse | null> = {
    message_start: () => handleMessageStart(ctx),
    content_block_start: () => handleContentBlockStart(event, ctx),
    content_block_delta: () => handleContentBlockDelta(event, ctx),
    content_block_stop: () => handleContentBlockStop(event, ctx),
  };

  return handlers[event.type]?.() ?? null;
}
```

## Proposed Structure

```
claude-proxy-sdk/
├── src/
│   ├── index.ts
│   ├── claude.ts              # Slim, uses handlers
│   ├── handlers/
│   │   ├── index.ts           # Export all handlers
│   │   ├── messageHandlers.ts # Handler map and dispatch
│   │   ├── systemHandler.ts
│   │   ├── assistantHandler.ts
│   │   ├── streamEventHandler.ts
│   │   ├── toolProgressHandler.ts
│   │   └── userHandler.ts
│   └── context/
│       └── ExecutionContext.ts
```

## Tasks

- [ ] Create `src/handlers/` directory
- [ ] Extract system message handler
- [ ] Extract assistant message handler
- [ ] Extract stream event handlers (with sub-handlers)
- [ ] Extract tool progress handler
- [ ] Extract user message handler
- [ ] Create handler map in `messageHandlers.ts`
- [ ] Update `claude.ts` to use handlers
- [ ] Add unit tests for each handler

## Acceptance Criteria

- [ ] No switch statement over 50 lines
- [ ] Each handler in separate file
- [ ] Each handler independently testable
- [ ] Handler map extensible for new types
- [ ] All existing functionality preserved

## References

- `ARCHITECTURE_REVIEW.md` section "1. Refactor `processMessage`" (Proxy SDK)
- Current file: `claude-proxy-sdk/src/claude.ts`

## Labels

```
priority: P2
type: refactoring
component: proxy
```
