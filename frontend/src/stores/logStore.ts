/**
 * Svelte store for managing application logs
 * Connects to backend log WebSocket for real-time updates
 */

import { writable, derived } from 'svelte/store';

export type LogLevel = 'info' | 'warning' | 'error';

export interface LogEntry {
  id: number;
  level: LogLevel;
  message: string;
  timestamp: string;
}

export interface LogState {
  entries: LogEntry[];
  status: LogLevel;
  isConnected: boolean;
  unreadCount: number;
}

const initialState: LogState = {
  entries: [],
  status: 'info',
  isConnected: false,
  unreadCount: 0,
};

function createLogStore() {
  const { subscribe, set, update } = writable<LogState>(initialState);

  let ws: WebSocket | null = null;
  let reconnectTimeout: ReturnType<typeof setTimeout> | null = null;

  function connect() {
    if (ws?.readyState === WebSocket.OPEN) return;

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws/logs`;

    ws = new WebSocket(wsUrl);

    ws.onopen = () => {
      update((state) => ({ ...state, isConnected: true }));
      // Fetch existing logs on connect
      fetchLogs();
    };

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);

        if (data.type === 'status') {
          update((state) => ({ ...state, status: data.status }));
        } else if (data.type === 'log') {
          update((state) => {
            const newStatus = getHigherLevel(state.status, data.entry.level);
            return {
              ...state,
              entries: [...state.entries.slice(-99), data.entry], // Keep last 100
              status: newStatus,
              unreadCount: state.unreadCount + 1,
            };
          });
        }
      } catch (e) {
        console.error('[LogStore] Failed to parse message:', e);
      }
    };

    ws.onclose = () => {
      update((state) => ({ ...state, isConnected: false }));
      // Reconnect after 5 seconds
      reconnectTimeout = setTimeout(connect, 5000);
    };

    ws.onerror = () => {
      ws?.close();
    };
  }

  function disconnect() {
    if (reconnectTimeout) {
      clearTimeout(reconnectTimeout);
      reconnectTimeout = null;
    }
    ws?.close();
    ws = null;
  }

  async function fetchLogs() {
    try {
      const response = await fetch('/api/logs');
      if (response.ok) {
        const data = await response.json();
        update((state) => ({
          ...state,
          entries: data.entries || [],
          status: data.status || 'info',
        }));
      }
    } catch (e) {
      console.error('[LogStore] Failed to fetch logs:', e);
    }
  }

  async function clearStatus() {
    try {
      await fetch('/api/logs/clear', { method: 'POST' });
      update((state) => ({
        ...state,
        status: 'info',
        unreadCount: 0,
      }));
    } catch (e) {
      console.error('[LogStore] Failed to clear status:', e);
    }
  }

  function markAsRead() {
    update((state) => ({ ...state, unreadCount: 0 }));
  }

  return {
    subscribe,
    connect,
    disconnect,
    clearStatus,
    markAsRead,
    reset: () => set(initialState),
  };
}

function getHigherLevel(current: LogLevel, incoming: LogLevel): LogLevel {
  const levels: Record<LogLevel, number> = { info: 0, warning: 1, error: 2 };
  return levels[incoming] > levels[current] ? incoming : current;
}

export const logStore = createLogStore();

// Derived stores for convenience
export const logEntries = derived(logStore, ($store) => $store.entries);
export const logStatus = derived(logStore, ($store) => $store.status);
export const hasWarningOrError = derived(
  logStore,
  ($store) => $store.status === 'warning' || $store.status === 'error'
);
export const unreadLogCount = derived(logStore, ($store) => $store.unreadCount);
