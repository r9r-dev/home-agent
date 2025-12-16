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
  role: 'user' | 'assistant';
  content: string;
  timestamp: Date;
  attachments?: MessageAttachment[];
}

export type ClaudeModel = 'haiku' | 'sonnet' | 'opus';

export interface ChatState {
  messages: Message[];
  currentSessionId: string | null;
  selectedModel: ClaudeModel;
  isConnected: boolean;
  isTyping: boolean;
  error: string | null;
}

const initialState: ChatState = {
  messages: [],
  currentSessionId: null,
  selectedModel: 'haiku',
  isConnected: false,
  isTyping: false,
  error: null,
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
    addMessage: (role: 'user' | 'assistant', content: string, attachments?: MessageAttachment[]) => {
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
          messages[messages.length - 1] = {
            ...lastMessage,
            content: lastMessage.content + chunk,
          };
        }

        return {
          ...state,
          messages,
        };
      });
    },

    /**
     * Clear all messages and reset model to default
     */
    clearMessages: () => {
      update((state) => ({
        ...state,
        messages: [],
        currentSessionId: null,
        selectedModel: 'haiku',
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
     */
    setTyping: (isTyping: boolean) => {
      update((state) => ({
        ...state,
        isTyping,
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
     * Reset the entire store
     */
    reset: () => {
      set(initialState);
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
