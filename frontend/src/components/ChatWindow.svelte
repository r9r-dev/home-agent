<script lang="ts" module>
  declare const __APP_VERSION__: string;
</script>

<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { chatStore, currentThinking, type ClaudeModel, type MessageAttachment } from '../stores/chatStore';
  import { websocketService, type MessageAttachment as WsAttachment } from '../services/websocket';
  import { fetchMessages, fetchSession, updateSessionModel, fetchSettings, updateSetting, fetchToolCalls, type Message as ApiMessage, type UploadedFile, type ToolCallRecord } from '../services/api';
  import MessageList from './MessageList.svelte';
  import InputBox from './InputBox.svelte';
  import Sidebar from './Sidebar.svelte';
  import SettingsDialog from './SettingsDialog.svelte';
  import MemoryDialog from './MemoryDialog.svelte';
  import { Badge } from "$lib/components/ui/badge";
  import * as Alert from "$lib/components/ui/alert";
  import * as Menubar from "$lib/components/ui/menubar";
  import Icon from "@iconify/svelte";

  // App version (injected by Vite from package.json)
  const APP_VERSION = __APP_VERSION__;

  // Sidebar reference for refreshing
  let sidebar: { refresh: () => void };

  // Dialog states
  let settingsDialogOpen = $state(false);
  let memoryDialogOpen = $state(false);

  // Model options
  const models: { value: ClaudeModel; label: string }[] = [
    { value: 'haiku', label: 'Haiku' },
    { value: 'sonnet', label: 'Sonnet' },
    { value: 'opus', label: 'Opus' },
  ];

  // Handle model change from menubar
  async function handleModelChange(model: ClaudeModel) {
    chatStore.setModel(model);
    // Update in database if we have a session
    const sessionId = $chatStore.currentSessionId;
    if (sessionId) {
      try {
        await updateSessionModel(sessionId, model);
      } catch (error) {
        console.error('Failed to update session model:', error);
      }
    }
  }

  // Handle thinking mode toggle
  async function handleThinkingToggle() {
    const newValue = !chatState.thinkingEnabled;
    chatStore.setThinkingEnabled(newValue);
    try {
      await updateSetting('thinking_enabled', newValue ? 'true' : 'false');
    } catch (error) {
      console.error('Failed to save thinking setting:', error);
    }
  }

  // Load thinking setting on mount
  async function loadThinkingSetting() {
    try {
      const settings = await fetchSettings();
      if (settings.thinking_enabled === 'true') {
        chatStore.setThinkingEnabled(true);
      }
    } catch (error) {
      console.error('Failed to load thinking setting:', error);
    }
  }

  // Subscribe to store
  let chatState = $derived($chatStore);

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

      case 'thinking':
        chatStore.setTyping(true);
        if (data.content) {
          chatStore.appendToThinking(data.content);
        }
        break;

      case 'done':
        chatStore.setTyping(false);
        // Keep thinking content visible after response completes
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

      case 'tool_start':
        if (data.tool) {
          chatStore.startToolCall(data.tool.tool_use_id, data.tool.tool_name, data.tool.input);
        }
        break;

      case 'tool_progress':
        if (data.tool) {
          chatStore.updateToolProgress(data.tool.tool_use_id, data.elapsedTimeSeconds || 0);
        }
        break;

      case 'tool_input_delta':
        if (data.tool && data.input_delta) {
          chatStore.appendToolInputDelta(data.tool.tool_use_id, data.input_delta);
        }
        break;

      case 'tool_result':
      case 'tool_error':
        if (data.tool) {
          // Pass the final input from the tool result (SDK sends complete input here)
          chatStore.completeToolCall(data.tool.tool_use_id, data.toolOutput || '', data.isError || data.type === 'tool_error', data.tool.input);
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
    if ((!content.trim() && (!attachments || attachments.length === 0)) || !chatState.isConnected) {
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

      // Clear previous thinking content before new message
      chatStore.clearThinking();

      chatStore.addMessage('user', content, storeAttachments);
      websocketService.sendMessage(content, chatState.currentSessionId || undefined, chatState.selectedModel, wsAttachments, chatState.thinkingEnabled);
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
      const [session, apiMessages, toolCallRecords] = await Promise.all([
        fetchSession(sessionId),
        fetchMessages(sessionId),
        fetchToolCalls(sessionId),
      ]);

      // Combine messages and tool calls to assign orderIndex based on chronological order
      const allItems: { type: 'message' | 'tool'; timestamp: Date; data: unknown }[] = [];

      apiMessages.forEach((msg: ApiMessage) => {
        allItems.push({
          type: 'message',
          timestamp: new Date(msg.created_at),
          data: msg,
        });
      });

      toolCallRecords.forEach((tc: ToolCallRecord) => {
        allItems.push({
          type: 'tool',
          timestamp: new Date(tc.created_at),
          data: tc,
        });
      });

      // Sort by timestamp to get chronological order
      allItems.sort((a, b) => a.timestamp.getTime() - b.timestamp.getTime());

      // Assign orderIndex based on sorted position
      const messages = allItems
        .filter((item) => item.type === 'message')
        .map((item) => {
          const msg = item.data as ApiMessage;
          const orderIndex = allItems.indexOf(item);
          return {
            id: String(msg.id),
            role: msg.role,
            content: msg.content,
            timestamp: new Date(msg.created_at),
            orderIndex,
          };
        });

      const toolCalls = allItems
        .filter((item) => item.type === 'tool')
        .map((item) => {
          const tc = item.data as ToolCallRecord;
          const orderIndex = allItems.indexOf(item);
          return {
            toolUseId: tc.tool_use_id,
            toolName: tc.tool_name,
            status: tc.status,
            startTime: new Date(tc.created_at),
            endTime: tc.completed_at ? new Date(tc.completed_at) : undefined,
            orderIndex,
            // Don't load input/output here - lazy loading
          };
        });

      chatStore.loadMessages(sessionId, messages, session.model as ClaudeModel);
      chatStore.loadToolCalls(toolCalls);
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
    chatStore.clearToolCalls();
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

    // Load saved settings
    loadThinkingSetting();

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
    currentSessionId={chatState.currentSessionId}
    onSelectSession={handleSelectSession}
    onNewConversation={handleNewConversation}
  />

  <div class="flex flex-col flex-1 min-w-0 min-h-0 bg-background">
    <header class="bg-background border-b border-border shrink-0">
      <div class="flex justify-between items-center px-8 py-3 max-w-[1400px] mx-auto w-full">
        <div class="text-xl font-medium tracking-tight">
          <span class="text-foreground">hal</span><span class="text-muted-foreground">fred</span>
        </div>

        <div class="flex items-center gap-4">
          <Menubar.Root>
            <!-- Claude Menu -->
            <Menubar.Menu>
              <Menubar.Trigger class="text-sm font-normal">Claude</Menubar.Trigger>
              <Menubar.Content>
                <!-- Model Sub-menu -->
                <Menubar.Sub>
                  <Menubar.SubTrigger>
                    <Icon icon="mynaui:cpu" class="size-4 mr-2" />
                    Modele
                    <span class="ml-auto text-xs text-muted-foreground capitalize">{chatState.selectedModel}</span>
                  </Menubar.SubTrigger>
                  <Menubar.SubContent>
                    <Menubar.RadioGroup value={chatState.selectedModel}>
                      {#each models as model}
                        <Menubar.RadioItem
                          value={model.value}
                          onclick={() => handleModelChange(model.value)}
                        >
                          {model.label}
                        </Menubar.RadioItem>
                      {/each}
                    </Menubar.RadioGroup>
                  </Menubar.SubContent>
                </Menubar.Sub>

                <!-- Mode Thinking (styled as toggle) -->
                <Menubar.Item
                  onSelect={(e) => {
                    e.preventDefault();
                    handleThinkingToggle();
                  }}
                  class="flex items-center justify-between"
                >
                  <span>Mode Thinking</span>
                  <span
                    class="ml-4 relative inline-flex h-5 w-9 shrink-0 cursor-pointer rounded-full transition-colors {chatState.thinkingEnabled ? 'bg-primary' : 'bg-muted'}"
                  >
                    <span
                      class="pointer-events-none block h-4 w-4 rounded-full bg-white shadow-sm ring-0 transition-transform mt-0.5 {chatState.thinkingEnabled ? 'translate-x-4 ml-0.5' : 'translate-x-0.5'}"
                    ></span>
                  </span>
                </Menubar.Item>

                <Menubar.Separator />

                <!-- Memory -->
                <Menubar.Item onclick={() => memoryDialogOpen = true}>
                  <Icon icon="mynaui:brain" class="size-4 mr-2" />
                  Gestion memoire
                </Menubar.Item>
              </Menubar.Content>
            </Menubar.Menu>

            <!-- Host Menu -->
            <Menubar.Menu>
              <Menubar.Trigger class="text-sm font-normal">Host</Menubar.Trigger>
              <Menubar.Content>
                <Menubar.Item disabled>
                  <Icon icon="mynaui:download" class="size-4 mr-2" />
                  Mettre a jour Claude Code
                </Menubar.Item>
              </Menubar.Content>
            </Menubar.Menu>

            <!-- Parametres Menu -->
            <Menubar.Menu>
              <Menubar.Trigger class="text-sm font-normal" onclick={() => settingsDialogOpen = true}>
                Parametres
              </Menubar.Trigger>
            </Menubar.Menu>
          </Menubar.Root>

          <Badge
            variant="outline"
            class="gap-2 px-3 py-1.5 bg-black text-white border-black"
          >
            <span class="w-1.5 h-1.5 rounded-full {connectionStatus.color === 'destructive' ? 'bg-red-500' : connectionStatus.color === 'default' ? 'bg-green-500' : 'bg-gray-400'}"></span>
            {connectionStatus.text}
          </Badge>
        </div>
      </div>
    </header>

    {#if chatState.error}
      <Alert.Root variant="destructive" class="rounded-none border-x-0 border-t-0">
        <Icon icon="mynaui:danger-circle" class="size-4" />
        <Alert.Description>{chatState.error}</Alert.Description>
      </Alert.Root>
    {/if}

    <MessageList messages={chatState.messages} isTyping={chatState.isTyping} />

    <InputBox bind:this={inputBox} onSend={handleSendMessage} disabled={!chatState.isConnected || chatState.isTyping} sessionId={chatState.currentSessionId} />

    <footer class="py-2 px-4 text-center text-[0.625rem] text-muted-foreground font-mono bg-background">
      v{APP_VERSION}
    </footer>
  </div>
</div>

<!-- Settings Dialog -->
<SettingsDialog bind:open={settingsDialogOpen} />

<!-- Memory Dialog -->
<MemoryDialog bind:open={memoryDialogOpen} />
