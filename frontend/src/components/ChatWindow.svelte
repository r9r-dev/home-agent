<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { chatStore } from '../stores/chatStore';
  import { websocketService } from '../services/websocket';
  import { fetchMessages, type Message as ApiMessage } from '../services/api';
  import MessageList from './MessageList.svelte';
  import InputBox from './InputBox.svelte';
  import Sidebar from './Sidebar.svelte';

  // App version
  const APP_VERSION = '0.4.0';

  // Sidebar reference for refreshing
  let sidebar: { refresh: () => void };

  // Subscribe to store
  let state = $chatStore;
  $: state = $chatStore;

  // Cleanup functions
  let unsubscribeMessage: (() => void) | null = null;
  let unsubscribeOpen: (() => void) | null = null;
  let unsubscribeClose: (() => void) | null = null;
  let unsubscribeError: (() => void) | null = null;

  /**
   * Handle incoming WebSocket messages
   */
  function handleWebSocketMessage(data: any) {
    console.log('[ChatWindow] Received message:', data);

    switch (data.type) {
      case 'chunk':
        // Streaming response chunk
        chatStore.setTyping(true);
        chatStore.appendToLastMessage(data.content);
        break;

      case 'done':
        // Response complete
        chatStore.setTyping(false);
        if (data.sessionId) {
          chatStore.setSessionId(data.sessionId);
        }
        break;

      case 'error':
        // Error from backend
        chatStore.setTyping(false);
        chatStore.setError(data.message || data.error || 'An error occurred');
        break;

      case 'session':
      case 'session_id':
        // Session information
        if (data.sessionId) {
          chatStore.setSessionId(data.sessionId);
        }
        break;

      default:
        console.warn('[ChatWindow] Unknown message type:', data.type);
    }
  }

  /**
   * Handle WebSocket connection open
   */
  function handleWebSocketOpen() {
    console.log('[ChatWindow] Connected to server');
    chatStore.setConnected(true);
    chatStore.setError(null);
  }

  /**
   * Handle WebSocket connection close
   */
  function handleWebSocketClose(event: CloseEvent) {
    console.log('[ChatWindow] Disconnected from server');
    chatStore.setConnected(false);
    chatStore.setTyping(false);

    if (event.code !== 1000) {
      // Not a normal closure
      chatStore.setError('Connection lost. Attempting to reconnect...');
    }
  }

  /**
   * Handle WebSocket errors
   */
  function handleWebSocketError(error: Event) {
    console.error('[ChatWindow] WebSocket error:', error);
    chatStore.setError('Connection error. Please check your network.');
  }

  // Reference to InputBox for focus management
  let inputBox: { focus: () => void };

  /**
   * Send a message
   */
  function handleSendMessage(content: string) {
    if (!content.trim() || !state.isConnected) {
      return;
    }

    try {
      // Add user message to chat
      chatStore.addMessage('user', content);

      // Send to server with session ID for continuity
      websocketService.sendMessage(content, state.currentSessionId || undefined);

      // Clear any previous errors
      chatStore.setError(null);
    } catch (error) {
      console.error('[ChatWindow] Failed to send message:', error);
      chatStore.setError('Failed to send message. Please try again.');
    }
  }

  // Re-focus input when response is complete
  $: if (!$chatStore.isTyping && inputBox) {
    inputBox.focus();
  }

  /**
   * Handle selecting a session from the sidebar
   */
  async function handleSelectSession(sessionId: string) {
    try {
      // Fetch messages for the session
      const apiMessages = await fetchMessages(sessionId);

      // Convert API messages to store format
      const messages = apiMessages.map((msg: ApiMessage) => ({
        id: String(msg.id),
        role: msg.role,
        content: msg.content,
        timestamp: new Date(msg.created_at),
      }));

      // Load messages into store
      chatStore.loadMessages(sessionId, messages);
      chatStore.setError(null);
    } catch (error) {
      console.error('[ChatWindow] Failed to load session:', error);
      chatStore.setError('Failed to load conversation');
    }
  }

  /**
   * Handle creating a new conversation
   */
  function handleNewConversation() {
    chatStore.clearMessages();
    if (inputBox) {
      inputBox.focus();
    }
  }

  // Refresh sidebar when a message is sent (new session might be created)
  $: if (!$chatStore.isTyping && $chatStore.messages.length > 0 && sidebar) {
    sidebar.refresh();
  }

  /**
   * Get connection status display - reactive based on store values
   */
  $: connectionStatus = (() => {
    if ($chatStore.error) {
      return { text: 'Erreur', color: 'var(--color-error)' };
    }
    if ($chatStore.isConnected) {
      return { text: 'Connecté', color: 'var(--color-success)' };
    }
    return { text: 'Connexion...', color: 'var(--color-warning)' };
  })();

  /**
   * Lifecycle: mount
   */
  onMount(() => {
    console.log('[ChatWindow] Mounting component');

    // Register WebSocket event handlers
    unsubscribeMessage = websocketService.onMessage(handleWebSocketMessage);
    unsubscribeOpen = websocketService.onOpen(handleWebSocketOpen);
    unsubscribeClose = websocketService.onClose(handleWebSocketClose);
    unsubscribeError = websocketService.onError(handleWebSocketError);

    // Connect to WebSocket
    websocketService.connect();
  });

  /**
   * Lifecycle: destroy
   */
  onDestroy(() => {
    console.log('[ChatWindow] Unmounting component');

    // Cleanup event handlers
    if (unsubscribeMessage) unsubscribeMessage();
    if (unsubscribeOpen) unsubscribeOpen();
    if (unsubscribeClose) unsubscribeClose();
    if (unsubscribeError) unsubscribeError();

    // Disconnect WebSocket
    websocketService.disconnect();
  });
</script>

<div class="app-container">
  <Sidebar
    bind:this={sidebar}
    currentSessionId={state.currentSessionId}
    onSelectSession={handleSelectSession}
    onNewConversation={handleNewConversation}
  />

  <div class="chat-window">
    <header class="chat-header">
      <div class="header-content">
        <div class="header-logo">
          <span class="logo-home">home</span><span class="logo-agent">agent</span>
        </div>
        <nav class="header-nav">
          <a href="#" class="nav-link">Machines</a>
          <a href="#" class="nav-link">Containers</a>
          <div class="status-badge">
            <span class="status-dot" style="background-color: {connectionStatus.color}"></span>
            <span class="status-label">{connectionStatus.text}</span>
          </div>
        </nav>
      </div>
    </header>

    {#if state.error}
      <div class="error-banner" role="alert">
        <span>{state.error}</span>
      </div>
    {/if}

    <MessageList messages={state.messages} isTyping={state.isTyping} />

    <InputBox bind:this={inputBox} onSend={handleSendMessage} disabled={!state.isConnected || state.isTyping} />

    <footer class="app-footer">
      <span>v{APP_VERSION}</span>
    </footer>
  </div>
</div>

<style>
  .app-container {
    display: flex;
    height: 100vh;
    overflow: hidden;
  }

  .chat-window {
    display: flex;
    flex-direction: column;
    flex: 1;
    min-width: 0;
    background: var(--color-bg-primary);
  }

  .chat-header {
    background: var(--color-bg-primary);
    border-bottom: 1px solid var(--color-border);
    flex-shrink: 0;
  }

  .header-content {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem 2rem;
    max-width: 1400px;
    margin: 0 auto;
    width: 100%;
  }

  .header-logo {
    font-size: 1.25rem;
    font-weight: 500;
    letter-spacing: -0.02em;
  }

  .logo-home {
    color: var(--color-text-primary);
  }

  .logo-agent {
    color: var(--color-text-secondary);
  }

  .header-nav {
    display: flex;
    align-items: center;
    gap: 2rem;
  }

  .nav-link {
    color: var(--color-text-secondary);
    font-size: 0.875rem;
    text-decoration: none;
    transition: color var(--transition-fast);
  }

  .nav-link:hover {
    color: var(--color-text-primary);
  }

  .status-badge {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.375rem 0.75rem;
    background: var(--color-bg-tertiary);
    border-radius: 4px;
    border: 1px solid var(--color-border);
  }

  .status-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
  }

  .status-label {
    font-size: 0.75rem;
    color: var(--color-text-secondary);
  }

  .error-banner {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0.75rem 1.5rem;
    background: rgba(239, 68, 68, 0.1);
    border-bottom: 1px solid var(--color-error);
    color: var(--color-error);
    font-size: 0.875rem;
  }

  .app-footer {
    padding: 0.5rem 1rem;
    text-align: center;
    font-size: 0.625rem;
    color: var(--color-text-tertiary);
    font-family: var(--font-family-mono);
    border-top: 1px solid var(--color-border);
    background: var(--color-bg-primary);
  }

  /* Responsive */
  @media (max-width: 768px) {
    .app-container :global(.sidebar) {
      display: none;
    }

    .header-content {
      padding: 0.875rem 1rem;
    }

    .header-logo {
      font-size: 1rem;
    }

    .header-nav {
      gap: 1rem;
    }

    .nav-link {
      font-size: 0.75rem;
    }

    .status-badge {
      padding: 0.25rem 0.5rem;
    }

    .status-label {
      font-size: 0.625rem;
    }
  }
</style>
