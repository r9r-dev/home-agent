# Split ChatWindow.svelte into focused components

**Priority:** P2 (Medium)
**Type:** Refactoring
**Component:** Frontend
**Estimated Effort:** Medium

## Summary

Refactor `ChatWindow.svelte` (603 lines) by extracting responsibilities into focused child components.

## Current Responsibilities

- WebSocket lifecycle management
- Session management
- Dialog state management (settings, memory, update, search, logs)
- Message sending
- Search integration
- Model/thinking settings
- Keyboard shortcuts

## Proposed Structure

```
frontend/src/components/
├── chat/
│   ├── ChatWindow.svelte     # Slim coordinator (~100 lines)
│   ├── ChatHeader.svelte     # Menubar, logo, indicators
│   ├── ChatContent.svelte    # Message list + input wrapper
│   ├── EmptyState.svelte     # Welcome screen
│   └── DialogManager.svelte  # Dialog state coordination
├── layout/
│   └── AppLayout.svelte      # Overall layout structure
```

## Example Refactored ChatWindow

```svelte
<script lang="ts">
  import { useWebSocket } from '../hooks/useWebSocket';
  import { useChatSession } from '../hooks/useChatSession';
  import ChatHeader from './ChatHeader.svelte';
  import ChatContent from './ChatContent.svelte';
  import DialogManager from './DialogManager.svelte';
  import Sidebar from './Sidebar.svelte';

  const { isConnected, sendMessage } = useWebSocket();
  const { session, loadSession, newConversation } = useChatSession();
</script>

<div class="flex h-screen overflow-hidden">
  <Sidebar {session} onSelect={loadSession} onNew={newConversation} />

  <main class="flex-1 flex flex-col">
    <ChatHeader />
    <ChatContent {isConnected} onSend={sendMessage} />
  </main>

  <DialogManager />
</div>
```

## Tasks

- [ ] Create `ChatHeader.svelte` with menubar and indicators
- [ ] Create `ChatContent.svelte` for messages + input
- [ ] Create `EmptyState.svelte` for welcome screen
- [ ] Create `DialogManager.svelte` for all dialogs
- [ ] Create custom hooks for WebSocket and session logic
- [ ] Refactor `ChatWindow.svelte` to coordinate children
- [ ] Ensure all functionality preserved
- [ ] Test on mobile breakpoints

## Acceptance Criteria

- [ ] `ChatWindow.svelte` under 150 lines
- [ ] Each component has single responsibility
- [ ] No breaking changes to functionality
- [ ] Mobile layout still works
- [ ] All dialogs still function correctly

## References

- `ARCHITECTURE_REVIEW.md` section "1. Split `ChatWindow.svelte`" (Frontend)
- Current file: `frontend/src/components/ChatWindow.svelte`

## Labels

```
priority: P2
type: refactoring
component: frontend
```
