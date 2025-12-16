<script lang="ts" module>
  declare const __APP_VERSION__: string;
</script>

<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { chatStore, type ClaudeModel, type MessageAttachment } from '../stores/chatStore';
  import { websocketService, type MessageAttachment as WsAttachment } from '../services/websocket';
  import { fetchMessages, fetchSession, type Message as ApiMessage, type UploadedFile } from '../services/api';
  import MessageList from './MessageList.svelte';
  import InputBox from './InputBox.svelte';
  import Sidebar from './Sidebar.svelte';
  import ModelSelector from './ModelSelector.svelte';
  import { Badge } from "$lib/components/ui/badge";
  import { Button } from "$lib/components/ui/button";
  import * as Alert from "$lib/components/ui/alert";
  import Icon from "@iconify/svelte";

  // App version (injected by Vite from package.json)
  const APP_VERSION = __APP_VERSION__;

  // Sidebar reference for refreshing
  let sidebar: { refresh: () => void };

  // Subscribe to store
  let state = $derived($chatStore);

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
        chatStore.setTyping(true);
        chatStore.appendToLastMessage(data.content);
        break;

      case 'done':
        chatStore.setTyping(false);
        if (data.sessionId) {
          chatStore.setSessionId(data.sessionId);
        }
        break;

      case 'error':
        chatStore.setTyping(false);
        chatStore.setError(data.message || data.error || 'An error occurred');
        break;

      case 'session':
      case 'session_id':
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
  function handleSendMessage(content: string, attachments?: UploadedFile[]) {
    if ((!content.trim() && (!attachments || attachments.length === 0)) || !state.isConnected) {
      return;
    }

    try {
      // Convert UploadedFile to MessageAttachment for store
      const storeAttachments: MessageAttachment[] | undefined = attachments?.map(a => ({
        id: a.id,
        filename: a.filename,
        path: a.path,
        type: a.type,
        mimeType: a.mime_type,
      }));

      // Convert to WebSocket format
      const wsAttachments: WsAttachment[] | undefined = attachments?.map(a => ({
        id: a.id,
        filename: a.filename,
        path: a.path,
        type: a.type,
        mime_type: a.mime_type,
      }));

      chatStore.addMessage('user', content, storeAttachments);
      websocketService.sendMessage(content, state.currentSessionId || undefined, state.selectedModel, wsAttachments);
      chatStore.setError(null);
    } catch (error) {
      console.error('[ChatWindow] Failed to send message:', error);
      chatStore.setError('Failed to send message. Please try again.');
    }
  }

  // Re-focus input when response is complete
  $effect(() => {
    if (!$chatStore.isTyping && inputBox) {
      inputBox.focus();
    }
  });

  /**
   * Handle selecting a session from the sidebar
   */
  async function handleSelectSession(sessionId: string) {
    try {
      const [session, apiMessages] = await Promise.all([
        fetchSession(sessionId),
        fetchMessages(sessionId),
      ]);

      const messages = apiMessages.map((msg: ApiMessage) => ({
        id: String(msg.id),
        role: msg.role,
        content: msg.content,
        timestamp: new Date(msg.created_at),
      }));

      chatStore.loadMessages(sessionId, messages, session.model as ClaudeModel);
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
  $effect(() => {
    if (!$chatStore.isTyping && $chatStore.messages.length > 0 && sidebar) {
      sidebar.refresh();
    }
  });

  /**
   * Get connection status display
   */
  let connectionStatus = $derived.by(() => {
    if ($chatStore.error) {
      return { text: 'Erreur', color: 'destructive' as const };
    }
    if ($chatStore.isConnected) {
      return { text: 'Connecté', color: 'default' as const };
    }
    return { text: 'Connexion...', color: 'secondary' as const };
  });

  /**
   * Lifecycle: mount
   */
  onMount(() => {
    console.log('[ChatWindow] Mounting component');

    unsubscribeMessage = websocketService.onMessage(handleWebSocketMessage);
    unsubscribeOpen = websocketService.onOpen(handleWebSocketOpen);
    unsubscribeClose = websocketService.onClose(handleWebSocketClose);
    unsubscribeError = websocketService.onError(handleWebSocketError);

    websocketService.connect();
  });

  /**
   * Lifecycle: destroy
   */
  onDestroy(() => {
    console.log('[ChatWindow] Unmounting component');

    if (unsubscribeMessage) unsubscribeMessage();
    if (unsubscribeOpen) unsubscribeOpen();
    if (unsubscribeClose) unsubscribeClose();
    if (unsubscribeError) unsubscribeError();

    websocketService.disconnect();
  });
</script>

<div class="flex h-screen overflow-hidden">
  <Sidebar
    bind:this={sidebar}
    currentSessionId={state.currentSessionId}
    onSelectSession={handleSelectSession}
    onNewConversation={handleNewConversation}
  />

  <div class="flex flex-col flex-1 min-w-0 min-h-0 bg-background">
    <header class="bg-background border-b border-border shrink-0">
      <div class="flex justify-between items-center px-8 py-4 max-w-[1400px] mx-auto w-full">
        <div class="text-xl font-medium tracking-tight">
          <span class="text-foreground">home</span><span class="text-muted-foreground">agent</span>
        </div>
        <nav class="flex items-center gap-8">
          <Button variant="ghost" class="text-muted-foreground hover:text-foreground text-sm">
            Machines
          </Button>
          <Button variant="ghost" class="text-muted-foreground hover:text-foreground text-sm">
            Containers
          </Button>
          <ModelSelector />
          <Badge
            variant="outline"
            class="gap-2 px-3 py-1.5 bg-black text-white border-black dark:bg-white dark:text-black dark:border-white"
          >
            <span class="w-1.5 h-1.5 rounded-full {connectionStatus.color === 'destructive' ? 'bg-red-500' : connectionStatus.color === 'default' ? 'bg-green-500' : 'bg-gray-400'}"></span>
            {connectionStatus.text}
          </Badge>
        </nav>
      </div>
    </header>

    {#if state.error}
      <Alert.Root variant="destructive" class="rounded-none border-x-0 border-t-0">
        <Icon icon="mynaui:danger-circle" class="size-4" />
        <Alert.Description>{state.error}</Alert.Description>
      </Alert.Root>
    {/if}

    <MessageList messages={state.messages} isTyping={state.isTyping} />

    <InputBox bind:this={inputBox} onSend={handleSendMessage} disabled={!state.isConnected || state.isTyping} sessionId={state.currentSessionId} />

    <footer class="py-2 px-4 text-center text-[0.625rem] text-muted-foreground font-mono bg-background">
      v{APP_VERSION}
    </footer>
  </div>
</div>
