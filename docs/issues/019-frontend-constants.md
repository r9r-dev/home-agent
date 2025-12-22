# Extract frontend constants

**Priority:** P2 (Medium)
**Type:** Refactoring
**Component:** Frontend
**Estimated Effort:** Low
**Good First Issue:** Yes

## Summary

Centralize magic strings and constants scattered throughout frontend components and stores.

## Examples of Magic Values

```typescript
// Model names scattered in components
const models = [
  { value: 'haiku', label: 'Haiku' },
  { value: 'sonnet', label: 'Sonnet' },
  { value: 'opus', label: 'Opus' },
];

// WebSocket message types
if (data.type === 'chunk') { ... }
if (data.type === 'thinking') { ... }
if (data.type === 'done') { ... }

// Tool status
status: 'running' | 'success' | 'error'

// Roles
role: 'user' | 'assistant' | 'thinking'
```

## Proposed Solution

```typescript
// src/constants/models.ts
export const CLAUDE_MODELS = {
  HAIKU: 'haiku',
  SONNET: 'sonnet',
  OPUS: 'opus',
} as const;

export type ClaudeModel = typeof CLAUDE_MODELS[keyof typeof CLAUDE_MODELS];

export const MODEL_OPTIONS: { value: ClaudeModel; label: string }[] = [
  { value: CLAUDE_MODELS.HAIKU, label: 'Haiku' },
  { value: CLAUDE_MODELS.SONNET, label: 'Sonnet' },
  { value: CLAUDE_MODELS.OPUS, label: 'Opus' },
];

export const DEFAULT_MODEL = CLAUDE_MODELS.HAIKU;
```

```typescript
// src/constants/websocket.ts
export const WS_MESSAGE_TYPES = {
  CHUNK: 'chunk',
  THINKING: 'thinking',
  THINKING_END: 'thinking_end',
  DONE: 'done',
  ERROR: 'error',
  SESSION_ID: 'session_id',
  SESSION_TITLE: 'session_title',
  TOOL_START: 'tool_start',
  TOOL_PROGRESS: 'tool_progress',
  TOOL_INPUT_DELTA: 'tool_input_delta',
  TOOL_RESULT: 'tool_result',
  TOOL_ERROR: 'tool_error',
} as const;

export type WsMessageType = typeof WS_MESSAGE_TYPES[keyof typeof WS_MESSAGE_TYPES];
```

```typescript
// src/constants/roles.ts
export const MESSAGE_ROLES = {
  USER: 'user',
  ASSISTANT: 'assistant',
  THINKING: 'thinking',
} as const;

export type MessageRole = typeof MESSAGE_ROLES[keyof typeof MESSAGE_ROLES];
```

```typescript
// src/constants/status.ts
export const TOOL_STATUS = {
  RUNNING: 'running',
  SUCCESS: 'success',
  ERROR: 'error',
} as const;

export type ToolStatus = typeof TOOL_STATUS[keyof typeof TOOL_STATUS];
```

```typescript
// src/constants/index.ts
export * from './models';
export * from './websocket';
export * from './roles';
export * from './status';
```

## Usage

```typescript
// Before
if (data.type === 'chunk') { ... }

// After
import { WS_MESSAGE_TYPES } from '../constants';
if (data.type === WS_MESSAGE_TYPES.CHUNK) { ... }
```

## Tasks

- [ ] Create `src/constants/` directory
- [ ] Create `models.ts` with model constants
- [ ] Create `websocket.ts` with message types
- [ ] Create `roles.ts` with role constants
- [ ] Create `status.ts` with status constants
- [ ] Create `index.ts` with re-exports
- [ ] Update `ChatWindow.svelte` to use constants
- [ ] Update `chatStore.ts` to use constants
- [ ] Update `websocket.ts` service to use constants
- [ ] Update all other components

## Acceptance Criteria

- [ ] No magic strings for models, roles, or statuses
- [ ] All constants have TypeScript types
- [ ] Constants are exported from single index
- [ ] Components use constants instead of strings

## References

- `ARCHITECTURE_REVIEW.md` section "4. Extract Constants"

## Notes

This is a good first issue for new contributors as it:
- Has clear, mechanical changes
- Improves code quality
- Provides good TypeScript practice
- Low risk of breaking changes

## Labels

```
priority: P2
type: refactoring
component: frontend
good first issue
```
