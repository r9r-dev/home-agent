/**
 * Svelte store for managing chat state
 * Handles messages, connection status, and typing indicators
 */

import { writable, derived } from 'svelte/store';

export interface MessageAttachment {
  id: string;
  filename: string;
  path: string;
  type: 'image' | 'file';
  mimeType?: string;
}

export interface Message {
  id: string;
  role: 'user' | 'assistant' | 'thinking';
  content: string;
  timestamp: Date;
  attachments?: MessageAttachment[];
}

export type ClaudeModel = 'haiku' | 'sonnet' | 'opus';

export type ToolCallStatus = 'running' | 'success' | 'error';

export interface ToolCall {
  toolUseId: string;
  toolName: string;
  input?: Record<string, unknown>;
  output?: string;
  status: ToolCallStatus;
  elapsedTimeSeconds?: number;
  startTime: Date;
  endTime?: Date;
}

export interface ChatState {
  messages: Message[];
  currentSessionId: string | null;
  selectedModel: ClaudeModel;
  isConnected: boolean;
  isTyping: boolean;
  error: string | null;
  responseCompleted: boolean; // Track if last response was completed (for paragraph separation)
  thinkingEnabled: boolean; // Extended thinking mode enabled
  currentThinking: string | null; // Current thinking content being streamed
  activeToolCalls: Map<string, ToolCall>; // Active tool calls keyed by toolUseId
}

// Initial state - sessionId is null until SDK provides one
const initialState: ChatState = {
  messages: [],
  currentSessionId: null,
  selectedModel: 'haiku',
  isConnected: false,
  isTyping: false,
  error: null,
  responseCompleted: false,
  thinkingEnabled: false,
  currentThinking: null,
  activeToolCalls: new Map(),
};

/**
 * Create the chat store
 */
function createChatStore() {
  const { subscribe, set, update } = writable<ChatState>(initialState);

  return {
    subscribe,

    /**
     * Add a new message to the chat
     */
    addMessage: (role: 'user' | 'assistant' | 'thinking', content: string, attachments?: MessageAttachment[]) => {
      update((state) => {
        const newMessage: Message = {
          id: crypto.randomUUID(),
          role,
          content,
          timestamp: new Date(),
          attachments,
        };
        return {
          ...state,
          messages: [...state.messages, newMessage],
        };
      });
    },

    /**
     * Update the last message (for streaming responses)
     */
    updateLastMessage: (content: string) => {
      update((state) => {
        if (state.messages.length === 0) {
          return state;
        }

        const messages = [...state.messages];
        const lastMessage = messages[messages.length - 1];

        if (lastMessage.role !== 'assistant') {
          // If last message is not from assistant, create a new one
          messages.push({
            id: crypto.randomUUID(),
            role: 'assistant',
            content,
            timestamp: new Date(),
          });
        } else {
          // Update existing assistant message
          messages[messages.length - 1] = {
            ...lastMessage,
            content,
          };
        }

        return {
          ...state,
          messages,
        };
      });
    },

    /**
     * Append content to the last message (for streaming)
     */
    appendToLastMessage: (chunk: string) => {
      update((state) => {
        if (state.messages.length === 0) {
          // Create new message if none exists
          const newMessage: Message = {
            id: crypto.randomUUID(),
            role: 'assistant',
            content: chunk,
            timestamp: new Date(),
          };
          return {
            ...state,
            messages: [newMessage],
            responseCompleted: false,
          };
        }

        const messages = [...state.messages];
        const lastMessage = messages[messages.length - 1];

        if (lastMessage.role !== 'assistant') {
          // Create new assistant message
          messages.push({
            id: crypto.randomUUID(),
            role: 'assistant',
            content: chunk,
            timestamp: new Date(),
          });
        } else {
          // Append to existing assistant message
          // Add paragraph separator if previous response was completed
          const separator = state.responseCompleted ? '\n\n' : '';
          messages[messages.length - 1] = {
            ...lastMessage,
            content: lastMessage.content + separator + chunk,
          };
        }

        return {
          ...state,
          messages,
          responseCompleted: false,
        };
      });
    },

    /**
     * Clear all messages and reset model to default
     * SessionId will be set when SDK provides one
     */
    clearMessages: () => {
      update((state) => ({
        ...state,
        messages: [],
        currentSessionId: null,
        selectedModel: 'haiku',
        currentThinking: null,
      }));
    },

    /**
     * Load messages from an existing session
     */
    loadMessages: (sessionId: string, loadedMessages: Message[], model?: ClaudeModel) => {
      update((state) => ({
        ...state,
        messages: loadedMessages,
        currentSessionId: sessionId,
        selectedModel: model || state.selectedModel,
      }));
    },

    /**
     * Set selected Claude model
     */
    setModel: (model: ClaudeModel) => {
      update((state) => ({
        ...state,
        selectedModel: model,
      }));
    },

    /**
     * Set connection status
     */
    setConnected: (isConnected: boolean) => {
      update((state) => ({
        ...state,
        isConnected,
        error: isConnected ? null : state.error,
      }));
    },

    /**
     * Set typing indicator
     * When typing stops, mark response as completed for paragraph separation
     */
    setTyping: (isTyping: boolean) => {
      update((state) => ({
        ...state,
        isTyping,
        responseCompleted: !isTyping ? true : state.responseCompleted,
      }));
    },

    /**
     * Set current session ID
     */
    setSessionId: (sessionId: string | null) => {
      update((state) => ({
        ...state,
        currentSessionId: sessionId,
      }));
    },

    /**
     * Set error message
     */
    setError: (error: string | null) => {
      update((state) => ({
        ...state,
        error,
      }));
    },

    /**
     * Set thinking mode enabled/disabled
     */
    setThinkingEnabled: (enabled: boolean) => {
      update((state) => ({
        ...state,
        thinkingEnabled: enabled,
      }));
    },

    /**
     * Append content to current thinking
     */
    appendToThinking: (chunk: string) => {
      update((state) => ({
        ...state,
        currentThinking: (state.currentThinking || '') + chunk,
      }));
    },

    /**
     * Clear current thinking content
     */
    clearThinking: () => {
      update((state) => ({
        ...state,
        currentThinking: null,
      }));
    },

    /**
     * Start a new tool call
     */
    startToolCall: (toolUseId: string, toolName: string, input?: Record<string, unknown>) => {
      update((state) => {
        const newMap = new Map(state.activeToolCalls);
        newMap.set(toolUseId, {
          toolUseId,
          toolName,
          input,
          status: 'running',
          startTime: new Date(),
        });
        return { ...state, activeToolCalls: newMap };
      });
    },

    /**
     * Update tool call progress (elapsed time)
     */
    updateToolProgress: (toolUseId: string, elapsedTimeSeconds: number) => {
      update((state) => {
        const newMap = new Map(state.activeToolCalls);
        const tool = newMap.get(toolUseId);
        if (tool) {
          newMap.set(toolUseId, { ...tool, elapsedTimeSeconds });
        }
        return { ...state, activeToolCalls: newMap };
      });
    },

    /**
     * Complete a tool call with output
     */
    completeToolCall: (toolUseId: string, output: string, isError: boolean) => {
      update((state) => {
        const newMap = new Map(state.activeToolCalls);
        const tool = newMap.get(toolUseId);
        if (tool) {
          newMap.set(toolUseId, {
            ...tool,
            output,
            status: isError ? 'error' : 'success',
            endTime: new Date(),
          });
        }
        return { ...state, activeToolCalls: newMap };
      });
    },

    /**
     * Load tool calls from an existing session
     */
    loadToolCalls: (toolCalls: ToolCall[]) => {
      update((state) => {
        const newMap = new Map<string, ToolCall>();
        for (const tc of toolCalls) {
          newMap.set(tc.toolUseId, tc);
        }
        return { ...state, activeToolCalls: newMap };
      });
    },

    /**
     * Clear all tool calls
     */
    clearToolCalls: () => {
      update((state) => ({
        ...state,
        activeToolCalls: new Map(),
      }));
    },

    /**
     * Reset the entire store
     * SessionId will be set when SDK provides one
     */
    reset: () => {
      set({
        ...initialState,
        activeToolCalls: new Map(), // Ensure new Map instance
      });
    },
  };
}

export const chatStore = createChatStore();

// Derived stores for convenience
export const messages = derived(chatStore, ($store) => $store.messages);
export const isConnected = derived(chatStore, ($store) => $store.isConnected);
export const isTyping = derived(chatStore, ($store) => $store.isTyping);
export const error = derived(chatStore, ($store) => $store.error);
export const selectedModel = derived(chatStore, ($store) => $store.selectedModel);
export const thinkingEnabled = derived(chatStore, ($store) => $store.thinkingEnabled);
export const currentThinking = derived(chatStore, ($store) => $store.currentThinking);
export const activeToolCalls = derived(chatStore, ($store) =>
  Array.from($store.activeToolCalls.values())
);
