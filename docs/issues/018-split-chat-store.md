# Split chatStore.ts into focused stores

**Priority:** P3 (Low)
**Type:** Refactoring
**Component:** Frontend
**Estimated Effort:** Medium

## Summary

Split the large `chatStore.ts` (544 lines) into smaller, domain-specific stores for better maintainability.

## Current State

Single store handling multiple concerns:
- Message state
- Session state
- Tool call state
- Thinking state
- Connection state
- Model selection

## Proposed Structure

```
frontend/src/stores/
├── messages.ts      # Message list state
├── session.ts       # Session ID and metadata
├── toolCalls.ts     # Tool call tracking
├── thinking.ts      # Thinking block state
├── connection.ts    # WebSocket connection state
├── chat.ts          # Composed store (facade)
└── index.ts         # Re-exports
```

## Implementation

### messages.ts

```typescript
// src/stores/messages.ts
import { writable, derived } from 'svelte/store';
import type { Message } from '../types';

function createMessagesStore() {
  const messages = writable<Message[]>([]);

  return {
    subscribe: messages.subscribe,

    add(role: string, content: string, attachments?: any[]) {
      messages.update(msgs => [...msgs, {
        id: crypto.randomUUID(),
        role,
        content,
        attachments,
        timestamp: new Date(),
        orderIndex: msgs.length,
      }]);
    },

    appendToLast(content: string) {
      messages.update(msgs => {
        if (msgs.length === 0) return msgs;
        const last = msgs[msgs.length - 1];
        return [...msgs.slice(0, -1), { ...last, content: last.content + content }];
      });
    },

    load(newMessages: Message[]) {
      messages.set(newMessages);
    },

    clear() {
      messages.set([]);
    },
  };
}

export const messagesStore = createMessagesStore();
```

### session.ts

```typescript
// src/stores/session.ts
import { writable } from 'svelte/store';
import type { ClaudeModel } from '../types';

function createSessionStore() {
  const sessionId = writable<string | null>(null);
  const model = writable<ClaudeModel>('haiku');

  return {
    sessionId: { subscribe: sessionId.subscribe },
    model: { subscribe: model.subscribe },

    setSessionId(id: string | null) {
      sessionId.set(id);
    },

    setModel(m: ClaudeModel) {
      model.set(m);
    },

    reset() {
      sessionId.set(null);
      model.set('haiku');
    },
  };
}

export const sessionStore = createSessionStore();
```

### toolCalls.ts

```typescript
// src/stores/toolCalls.ts
import { writable, derived } from 'svelte/store';
import type { ToolCall } from '../types';

function createToolCallsStore() {
  const toolCalls = writable<Map<string, ToolCall>>(new Map());

  return {
    subscribe: derived(toolCalls, $tc => Array.from($tc.values())).subscribe,

    start(toolUseId: string, toolName: string, input: any) {
      toolCalls.update(map => {
        map.set(toolUseId, {
          toolUseId,
          toolName,
          input,
          status: 'running',
          startTime: new Date(),
          orderIndex: map.size,
        });
        return new Map(map);
      });
    },

    updateProgress(toolUseId: string, elapsedSeconds: number) {
      toolCalls.update(map => {
        const tc = map.get(toolUseId);
        if (tc) {
          map.set(toolUseId, { ...tc, elapsedSeconds });
        }
        return new Map(map);
      });
    },

    complete(toolUseId: string, output: string, isError: boolean) {
      toolCalls.update(map => {
        const tc = map.get(toolUseId);
        if (tc) {
          map.set(toolUseId, {
            ...tc,
            output,
            status: isError ? 'error' : 'success',
            endTime: new Date(),
          });
        }
        return new Map(map);
      });
    },

    clear() {
      toolCalls.set(new Map());
    },
  };
}

export const toolCallsStore = createToolCallsStore();
```

### chat.ts (Facade)

```typescript
// src/stores/chat.ts
import { derived } from 'svelte/store';
import { messagesStore } from './messages';
import { sessionStore } from './session';
import { toolCallsStore } from './toolCalls';
import { thinkingStore } from './thinking';
import { connectionStore } from './connection';

// Composed store for backwards compatibility
export const chatStore = {
  // Expose individual stores
  messages: messagesStore,
  session: sessionStore,
  toolCalls: toolCallsStore,
  thinking: thinkingStore,
  connection: connectionStore,

  // Derived unified state (for components that need everything)
  state: derived(
    [messagesStore, sessionStore.sessionId, sessionStore.model, connectionStore],
    ([$messages, $sessionId, $model, $connection]) => ({
      messages: $messages,
      currentSessionId: $sessionId,
      selectedModel: $model,
      isConnected: $connection.isConnected,
      isTyping: $connection.isTyping,
    })
  ),

  // Facade methods
  addMessage: messagesStore.add,
  setSessionId: sessionStore.setSessionId,
  setModel: sessionStore.setModel,

  reset() {
    messagesStore.clear();
    sessionStore.reset();
    toolCallsStore.clear();
    thinkingStore.clear();
  },
};
```

## Tasks

- [ ] Create `messages.ts` store
- [ ] Create `session.ts` store
- [ ] Create `toolCalls.ts` store
- [ ] Create `thinking.ts` store
- [ ] Create `connection.ts` store
- [ ] Create `chat.ts` facade for backwards compatibility
- [ ] Update component imports gradually
- [ ] Remove old `chatStore.ts` after migration
- [ ] Ensure reactivity preserved

## Acceptance Criteria

- [ ] Each store under 100 lines
- [ ] Stores are independently usable
- [ ] Backwards compatibility via facade
- [ ] All tests pass
- [ ] No breaking changes to components

## References

- `ARCHITECTURE_REVIEW.md` section "5. Split `chatStore.ts`"
- Current file: `frontend/src/stores/chatStore.ts`

## Labels

```
priority: P3
type: refactoring
component: frontend
```
