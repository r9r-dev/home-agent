# Create shared frontend type definitions

**Priority:** P1 (High)
**Type:** Refactoring
**Component:** Frontend
**Estimated Effort:** Low

## Summary

Create shared TypeScript types to eliminate duplication between stores, services, and components.

## Current State

Types are duplicated:
- `MessageAttachment` in `chatStore.ts`
- `MessageAttachment` in `websocket.ts`
- `Attachment` in API types

## Proposed Structure

```typescript
// src/types/index.ts
export type AttachmentType = 'image' | 'file';

export interface Attachment {
  id: string;
  filename: string;
  path: string;
  type: AttachmentType;
  mimeType?: string;
}

export type MessageRole = 'user' | 'assistant' | 'thinking';
export type ClaudeModel = 'haiku' | 'sonnet' | 'opus';
export type ToolCallStatus = 'running' | 'success' | 'error';

export interface Message {
  id: string;
  role: MessageRole;
  content: string;
  timestamp: Date;
  attachments?: Attachment[];
}

// src/types/api.ts
export interface ApiSession {
  id: number;
  session_id: string;
  claude_session_id: string;
  title: string;
  model: ClaudeModel;
  created_at: string;
  last_activity: string;
}

export interface ApiMessage {
  id: number;
  session_id: string;
  role: MessageRole;
  content: string;
  created_at: string;
}

// Type guard functions
export function isApiSession(obj: unknown): obj is ApiSession {
  return typeof obj === 'object' && obj !== null && 'session_id' in obj;
}
```

## Tasks

- [ ] Create `src/types/` directory
- [ ] Define shared `Attachment` type
- [ ] Define shared `Message` type
- [ ] Define model and role constants
- [ ] Create API response types in `types/api.ts`
- [ ] Update `chatStore.ts` to use shared types
- [ ] Update `websocket.ts` to use shared types
- [ ] Update `api.ts` to use shared types
- [ ] Add type guard functions

## Acceptance Criteria

- [ ] Single source of truth for frontend types
- [ ] All stores and services use shared types
- [ ] No `any` types where specific types exist
- [ ] Type guards for runtime validation

## References

- `ARCHITECTURE_REVIEW.md` section "3. Consolidate Type Definitions" (Frontend)

## Labels

```
priority: P1
type: refactoring
component: frontend
```
