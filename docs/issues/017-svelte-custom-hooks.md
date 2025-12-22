# Create custom Svelte hooks for logic extraction

**Priority:** P2 (Medium)
**Type:** Refactoring
**Component:** Frontend
**Estimated Effort:** Medium

## Summary

Extract component logic into reusable Svelte hooks (composable functions), improving testability and reusability.

## Current State

Logic is embedded directly in `ChatWindow.svelte`:
- WebSocket connection management
- Session loading and state
- Keyboard shortcuts
- Settings persistence

## Proposed Hooks

### useWebSocket

```typescript
// src/hooks/useWebSocket.ts
import { writable, type Readable } from 'svelte/store';
import { onMount, onDestroy } from 'svelte';
import { websocketService } from '../services/websocket';

interface UseWebSocketReturn {
  isConnected: Readable<boolean>;
  sendMessage: typeof websocketService.sendMessage;
  onMessage: typeof websocketService.onMessage;
}

export function useWebSocket(): UseWebSocketReturn {
  const connected = writable(false);

  let unsubOpen: (() => void) | null = null;
  let unsubClose: (() => void) | null = null;

  onMount(() => {
    unsubOpen = websocketService.onOpen(() => connected.set(true));
    unsubClose = websocketService.onClose(() => connected.set(false));
    websocketService.connect();
  });

  onDestroy(() => {
    unsubOpen?.();
    unsubClose?.();
    websocketService.disconnect();
  });

  return {
    isConnected: { subscribe: connected.subscribe },
    sendMessage: websocketService.sendMessage.bind(websocketService),
    onMessage: websocketService.onMessage.bind(websocketService),
  };
}
```

### useChatSession

```typescript
// src/hooks/useChatSession.ts
import { chatStore } from '../stores/chatStore';
import { fetchSession, fetchMessages, fetchToolCalls } from '../services/api';

export function useChatSession() {
  async function loadSession(sessionId: string) {
    const [session, messages, toolCalls] = await Promise.all([
      fetchSession(sessionId),
      fetchMessages(sessionId),
      fetchToolCalls(sessionId),
    ]);

    // Process and load into store
    const processedMessages = processMessages(messages, toolCalls);
    chatStore.loadMessages(sessionId, processedMessages, session.model);
    chatStore.loadToolCalls(toolCalls);
  }

  function newConversation() {
    chatStore.clearMessages();
    chatStore.clearToolCalls();
  }

  function reset() {
    chatStore.reset();
  }

  return {
    loadSession,
    newConversation,
    reset,
  };
}
```

### useKeyboardShortcuts

```typescript
// src/hooks/useKeyboardShortcuts.ts
import { onMount, onDestroy } from 'svelte';

type ShortcutHandler = () => void;

interface Shortcut {
  key: string;
  ctrl?: boolean;
  meta?: boolean;
  shift?: boolean;
  handler: ShortcutHandler;
}

export function useKeyboardShortcuts(shortcuts: Shortcut[]) {
  function handleKeydown(event: KeyboardEvent) {
    for (const shortcut of shortcuts) {
      const ctrlMatch = shortcut.ctrl ? event.ctrlKey : true;
      const metaMatch = shortcut.meta ? event.metaKey : true;
      const shiftMatch = shortcut.shift ? event.shiftKey : !event.shiftKey;

      if (
        event.key.toLowerCase() === shortcut.key.toLowerCase() &&
        ctrlMatch &&
        metaMatch &&
        shiftMatch
      ) {
        event.preventDefault();
        shortcut.handler();
        return;
      }
    }
  }

  onMount(() => {
    document.addEventListener('keydown', handleKeydown);
  });

  onDestroy(() => {
    document.removeEventListener('keydown', handleKeydown);
  });
}

// Usage in component
useKeyboardShortcuts([
  { key: 'k', meta: true, handler: () => searchDialogOpen = true },
  { key: 'n', meta: true, handler: () => handleNewConversation() },
]);
```

### useSettings

```typescript
// src/hooks/useSettings.ts
import { writable } from 'svelte/store';
import { onMount } from 'svelte';
import { fetchSettings, updateSetting } from '../services/api';

export function useSettings() {
  const loading = writable(true);
  const thinkingEnabled = writable(false);
  const customInstructions = writable('');

  onMount(async () => {
    try {
      const settings = await fetchSettings();
      thinkingEnabled.set(settings.thinking_enabled === 'true');
      customInstructions.set(settings.custom_instructions || '');
    } finally {
      loading.set(false);
    }
  });

  async function setThinkingEnabled(enabled: boolean) {
    thinkingEnabled.set(enabled);
    await updateSetting('thinking_enabled', enabled ? 'true' : 'false');
  }

  return {
    loading: { subscribe: loading.subscribe },
    thinkingEnabled: { subscribe: thinkingEnabled.subscribe },
    customInstructions: { subscribe: customInstructions.subscribe },
    setThinkingEnabled,
  };
}
```

## Tasks

- [ ] Create `src/hooks/` directory
- [ ] Implement `useWebSocket` hook
- [ ] Implement `useChatSession` hook
- [ ] Implement `useKeyboardShortcuts` hook
- [ ] Implement `useSettings` hook
- [ ] Update `ChatWindow.svelte` to use hooks
- [ ] Document hook patterns and usage
- [ ] Add examples to component documentation

## Acceptance Criteria

- [ ] Hooks are reusable across components
- [ ] Logic is decoupled from UI
- [ ] Hooks handle cleanup properly (onDestroy)
- [ ] TypeScript types are complete
- [ ] Existing functionality unchanged

## References

- `ARCHITECTURE_REVIEW.md` section "2. Create Custom Hooks"
- Svelte 5 runes: https://svelte.dev/docs/svelte/what-are-runes

## Labels

```
priority: P2
type: refactoring
component: frontend
```
