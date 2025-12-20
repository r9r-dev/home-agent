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
  orderIndex: number; // Global order index for chronological display
  attachments?: MessageAttachment[];
}

export type ClaudeModel = 'haiku' | 'sonnet' | 'opus';

export type ToolCallStatus = 'running' | 'success' | 'error';

export interface ToolCall {
  toolUseId: string;
  toolName: string;
  input?: Record<string, unknown>;
  inputJson?: string; // Raw JSON string for streaming display
  output?: string;
  status: ToolCallStatus;
  elapsedTimeSeconds?: number;
  startTime: Date;
  endTime?: Date;
  orderIndex: number; // Global order index for chronological display
}

// Flow item types for unified message + tool call display
export type FlowItemType = 'message' | 'tool_call' | 'thinking';

export interface FlowItem {
  type: FlowItemType;
  timestamp: Date;
  orderIndex: number;
  message?: Message;
  toolCall?: ToolCall;
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
  currentThinkingOrderIndex: number | null; // Order index for current thinking block
  activeToolCalls: Map<string, ToolCall>; // Active tool calls keyed by toolUseId
  orderCounter: number; // Global counter for ordering events chronologically
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
  currentThinkingOrderIndex: null,
  activeToolCalls: new Map(),
  orderCounter: 0,
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
          orderIndex: state.orderCounter,
          attachments,
        };
        return {
          ...state,
          messages: [...state.messages, newMessage],
          orderCounter: state.orderCounter + 1,
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
            orderIndex: state.orderCounter,
          });
          return {
            ...state,
            messages,
            orderCounter: state.orderCounter + 1,
          };
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
            orderIndex: state.orderCounter,
          };
          return {
            ...state,
            messages: [newMessage],
            responseCompleted: false,
            orderCounter: state.orderCounter + 1,
          };
        }

        const messages = [...state.messages];
        const lastMessage = messages[messages.length - 1];

        // Check if there's a tool call with higher orderIndex than last assistant message
        // If so, we need to create a new message instead of appending
        const lastAssistantMessage = [...messages].reverse().find(m => m.role === 'assistant');
        const hasToolCallAfterLastAssistant = lastAssistantMessage
          ? Array.from(state.activeToolCalls.values()).some(
              tc => tc.orderIndex > lastAssistantMessage.orderIndex
            )
          : false;

        if (lastMessage.role !== 'assistant' || hasToolCallAfterLastAssistant) {
          // Create new assistant message
          messages.push({
            id: crypto.randomUUID(),
            role: 'assistant',
            content: chunk,
            timestamp: new Date(),
            orderIndex: state.orderCounter,
          });
          return {
            ...state,
            messages,
            responseCompleted: false,
            orderCounter: state.orderCounter + 1,
          };
        } else {
          // Append to existing assistant message
          // Add paragraph separator if previous response was completed
          const separator = state.responseCompleted ? '\n\n' : '';
          messages[messages.length - 1] = {
            ...lastMessage,
            content: lastMessage.content + separator + chunk,
          };
          return {
            ...state,
            messages,
            responseCompleted: false,
          };
        }
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
        currentThinkingOrderIndex: null,
        orderCounter: 0,
      }));
    },

    /**
     * Load messages from an existing session
     */
    loadMessages: (sessionId: string, loadedMessages: Message[], model?: ClaudeModel) => {
      update((state) => {
        // Calculate the next orderCounter based on max orderIndex
        const maxOrderIndex = loadedMessages.reduce(
          (max, msg) => Math.max(max, msg.orderIndex ?? 0),
          0
        );
        return {
          ...state,
          messages: loadedMessages,
          currentSessionId: sessionId,
          selectedModel: model || state.selectedModel,
          orderCounter: Math.max(state.orderCounter, maxOrderIndex + 1),
        };
      });
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
     * Captures orderIndex when starting a new thinking block
     */
    appendToThinking: (chunk: string) => {
      update((state) => {
        // If this is the start of a new thinking block, capture the orderIndex
        const isNewBlock = state.currentThinking === null;
        return {
          ...state,
          currentThinking: (state.currentThinking || '') + chunk,
          currentThinkingOrderIndex: isNewBlock ? state.orderCounter : state.currentThinkingOrderIndex,
          orderCounter: isNewBlock ? state.orderCounter + 1 : state.orderCounter,
        };
      });
    },

    /**
     * Clear current thinking content
     */
    clearThinking: () => {
      update((state) => ({
        ...state,
        currentThinking: null,
        currentThinkingOrderIndex: null,
      }));
    },

    /**
     * Finalize current thinking content by converting it to a message
     * Called when thinking_end is received from backend
     */
    finalizeThinking: () => {
      update((state) => {
        if (!state.currentThinking) {
          return state;
        }

        // Create a new thinking message from the accumulated content
        const thinkingMessage: Message = {
          id: crypto.randomUUID(),
          role: 'thinking',
          content: state.currentThinking,
          timestamp: new Date(),
          orderIndex: state.currentThinkingOrderIndex ?? state.orderCounter,
        };

        return {
          ...state,
          messages: [...state.messages, thinkingMessage],
          currentThinking: null,
          currentThinkingOrderIndex: null,
          // Don't increment orderCounter here since we already captured it when thinking started
        };
      });
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
          orderIndex: state.orderCounter,
        });
        return { ...state, activeToolCalls: newMap, orderCounter: state.orderCounter + 1 };
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
     * Append input delta to a tool call (for streaming input display)
     */
    appendToolInputDelta: (toolUseId: string, delta: string) => {
      update((state) => {
        const newMap = new Map(state.activeToolCalls);
        const tool = newMap.get(toolUseId);
        if (tool) {
          const currentInput = tool.inputJson || '';
          newMap.set(toolUseId, { ...tool, inputJson: currentInput + delta });
        }
        return { ...state, activeToolCalls: newMap };
      });
    },

    /**
     * Complete a tool call with output and final input
     */
    completeToolCall: (toolUseId: string, output: string, isError: boolean, finalInput?: Record<string, unknown>) => {
      update((state) => {
        const newMap = new Map(state.activeToolCalls);
        const tool = newMap.get(toolUseId);
        if (tool) {
          // Update input if provided and current input is empty
          const updatedInput = (finalInput && Object.keys(finalInput).length > 0) ? finalInput : tool.input;
          // Also update inputJson if we have finalInput and inputJson is empty
          const updatedInputJson = (finalInput && Object.keys(finalInput).length > 0 && !tool.inputJson)
            ? JSON.stringify(finalInput, null, 2)
            : tool.inputJson;

          newMap.set(toolUseId, {
            ...tool,
            input: updatedInput,
            inputJson: updatedInputJson,
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
        let maxOrderIndex = state.orderCounter;
        for (const tc of toolCalls) {
          newMap.set(tc.toolUseId, tc);
          if (tc.orderIndex !== undefined) {
            maxOrderIndex = Math.max(maxOrderIndex, tc.orderIndex + 1);
          }
        }
        return { ...state, activeToolCalls: newMap, orderCounter: maxOrderIndex };
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

// Unified flow combining messages and ALL tool calls in chronological order by orderIndex
// This includes running tool calls inline (not separately at the bottom)
export const unifiedFlow = derived(chatStore, ($store): FlowItem[] => {
  const items: FlowItem[] = [];

  // Add messages
  for (const msg of $store.messages) {
    items.push({
      type: msg.role === 'thinking' ? 'thinking' : 'message',
      timestamp: msg.timestamp,
      orderIndex: msg.orderIndex,
      message: msg,
    });
  }

  // Add ALL tool calls (including running ones) - they will appear inline
  for (const toolCall of $store.activeToolCalls.values()) {
    items.push({
      type: 'tool_call',
      timestamp: toolCall.startTime,
      orderIndex: toolCall.orderIndex,
      toolCall,
    });
  }

  // Sort by orderIndex for correct chronological order
  items.sort((a, b) => a.orderIndex - b.orderIndex);

  return items;
});

// Running tool calls (for status tracking, but they're now displayed inline via unifiedFlow)
export const runningToolCalls = derived(chatStore, ($store) =>
  Array.from($store.activeToolCalls.values()).filter(tc => tc.status === 'running')
);

// Current thinking order index (for positioning streaming thinking correctly)
export const currentThinkingOrderIndex = derived(chatStore, ($store) => $store.currentThinkingOrderIndex);
