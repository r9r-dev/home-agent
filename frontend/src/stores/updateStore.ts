import { writable, derived } from 'svelte/store';

export type UpdateStatus = 'idle' | 'checking' | 'available' | 'updating' | 'reconnecting' | 'success' | 'error';

export interface LogEntry {
  id: number;
  timestamp: string;
  message: string;
  level: 'info' | 'warning' | 'error';
  source?: 'backend' | 'proxy' | 'system';
}

export interface VersionInfo {
  current: string;
  latest: string | null;
  updateAvailable: boolean;
}

export interface ApiUpdateStatus {
  backend: VersionInfo;
  proxy: VersionInfo;
}

interface UpdateState {
  status: UpdateStatus;
  backend: VersionInfo;
  proxy: VersionInfo;
  backendLogs: LogEntry[];
  proxyLogs: LogEntry[];
  isChecking: boolean;
  isUpdating: boolean;
  isReconnecting: boolean;
  error: string | null;
  updateAvailable: boolean;
}

const initialState: UpdateState = {
  status: 'idle',
  backend: {
    current: 'unknown',
    latest: null,
    updateAvailable: false,
  },
  proxy: {
    current: 'unknown',
    latest: null,
    updateAvailable: false,
  },
  backendLogs: [],
  proxyLogs: [],
  isChecking: false,
  isUpdating: false,
  isReconnecting: false,
  error: null,
  updateAvailable: false,
};

function createUpdateStore() {
  const { subscribe, update, set } = writable<UpdateState>(initialState);

  let logIdCounter = 0;
  let updateWs: WebSocket | null = null;
  let pollingInterval: ReturnType<typeof setInterval> | null = null;
  let backendUpdateSucceeded = false;
  let proxyNeedsUpdate = false;

  // Helper to get current state synchronously
  function getState(): UpdateState {
    let state = initialState;
    const unsubscribe = subscribe(s => { state = s; });
    unsubscribe();
    return state;
  }

  // Get WebSocket URL from current location
  function getWsUrl(): string {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    return `${protocol}//${window.location.host}/ws/update`;
  }

  // Create a system log entry
  function createSystemLog(message: string, level: LogEntry['level'] = 'info'): LogEntry {
    return {
      id: ++logIdCounter,
      timestamp: new Date().toISOString(),
      message,
      level,
      source: 'system',
    };
  }

  // Add a log entry to the appropriate panel
  function addLog(entry: LogEntry) {
    update(s => {
      if (entry.source === 'proxy') {
        return { ...s, proxyLogs: [...s.proxyLogs, entry] };
      } else {
        return { ...s, backendLogs: [...s.backendLogs, entry] };
      }
    });
  }

  // Poll backend health until it's available
  async function pollBackendHealth(): Promise<void> {
    return new Promise((resolve) => {
      let attempts = 0;
      const maxAttempts = 60; // 60 attempts x 2s = 2 minutes max
      const pollIntervalMs = 2000;

      update(s => ({
        ...s,
        isReconnecting: true,
        status: 'reconnecting',
      }));

      addLog(createSystemLog('Backend en cours de redemarrage, attente...'));

      pollingInterval = setInterval(async () => {
        attempts++;

        try {
          const response = await fetch('/api/health', {
            method: 'GET',
            cache: 'no-store',
          });

          if (response.ok) {
            // Backend is back!
            if (pollingInterval) {
              clearInterval(pollingInterval);
              pollingInterval = null;
            }

            addLog(createSystemLog('Backend disponible, reconnexion...'));

            update(s => ({
              ...s,
              isReconnecting: false,
              backend: { ...s.backend, updateAvailable: false },
            }));

            resolve();
          }
        } catch {
          // Backend not ready yet
          if (attempts % 5 === 0) {
            addLog(createSystemLog(`Tentative ${attempts}/${maxAttempts}...`, 'warning'));
          }
        }

        if (attempts >= maxAttempts) {
          if (pollingInterval) {
            clearInterval(pollingInterval);
            pollingInterval = null;
          }

          addLog(createSystemLog('Timeout: backend non disponible', 'error'));

          update(s => ({
            ...s,
            isReconnecting: false,
            isUpdating: false,
            status: 'error',
            error: 'Backend non disponible apres le redemarrage',
          }));

          resolve();
        }
      }, pollIntervalMs);
    });
  }

  // Connect to update WebSocket for log streaming
  function connectWebSocket(autoTriggerProxy = false): Promise<void> {
    return new Promise((resolve, reject) => {
      if (updateWs && updateWs.readyState === WebSocket.OPEN) {
        resolve();
        return;
      }

      updateWs = new WebSocket(getWsUrl());

      updateWs.onopen = () => {
        console.log('[UpdateStore] WebSocket connected');
        resolve();
      };

      updateWs.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);

          if (data.type === 'update_log') {
            const entry: LogEntry = {
              id: ++logIdCounter,
              timestamp: data.timestamp,
              message: data.message,
              level: data.level || 'info',
              source: data.source,
            };

            update(s => {
              if (data.source === 'proxy') {
                return { ...s, proxyLogs: [...s.proxyLogs, entry] };
              } else {
                return { ...s, backendLogs: [...s.backendLogs, entry] };
              }
            });
          } else if (data.type === 'update_status') {
            if (data.status === 'success') {
              if (data.target === 'backend') {
                backendUpdateSucceeded = true;
                update(s => ({
                  ...s,
                  backend: { ...s.backend, updateAvailable: false },
                }));
              } else if (data.target === 'proxy') {
                update(s => ({
                  ...s,
                  isUpdating: false,
                  status: 'success',
                  proxy: { ...s.proxy, updateAvailable: false },
                  updateAvailable: false,
                }));
              }
            } else if (data.status === 'error') {
              update(s => ({
                ...s,
                isUpdating: false,
                status: 'error',
                error: data.error || 'Update failed',
              }));
            }
          } else if (data.type === 'error') {
            update(s => ({
              ...s,
              error: data.error,
            }));
          }
        } catch (e) {
          console.error('[UpdateStore] Failed to parse message:', e);
        }
      };

      updateWs.onerror = (error) => {
        console.error('[UpdateStore] WebSocket error:', error);
        reject(error);
      };

      updateWs.onclose = async () => {
        console.log('[UpdateStore] WebSocket closed');
        updateWs = null;

        // If we were updating and backend update was triggered, start polling
        const currentState = getState();

        if (currentState.isUpdating && !currentState.isReconnecting) {
          addLog(createSystemLog('Connexion perdue pendant la mise a jour', 'warning'));

          // Poll until backend is back
          await pollBackendHealth();

          // Only proceed if we're still in updating state (no error occurred)
          const stateAfterPoll = getState();

          if (stateAfterPoll.status !== 'error') {
            // Reconnect WebSocket
            try {
              await connectWebSocket(false);

              // If proxy needs update, trigger it now
              if (proxyNeedsUpdate) {
                addLog(createSystemLog('Lancement de la mise a jour du proxy...'));
                await triggerProxyUpdate();
              } else {
                // No proxy update needed, we're done
                update(s => ({
                  ...s,
                  isUpdating: false,
                  status: 'success',
                }));
              }
            } catch (err) {
              addLog(createSystemLog('Echec de la reconnexion WebSocket', 'error'));
              update(s => ({
                ...s,
                isUpdating: false,
                status: 'error',
                error: 'Echec de la reconnexion',
              }));
            }
          }
        }
      };

      // Timeout after 5 seconds
      setTimeout(() => {
        if (updateWs && updateWs.readyState === WebSocket.CONNECTING) {
          updateWs.close();
          reject(new Error('WebSocket connection timeout'));
        }
      }, 5000);
    });
  }

  // Trigger proxy update via API
  async function triggerProxyUpdate(): Promise<void> {
    try {
      const response = await fetch('/api/update/proxy', { method: 'POST' });
      if (!response.ok) {
        throw new Error(`Proxy update failed: ${response.statusText}`);
      }
      // Logs will stream via WebSocket
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Proxy update failed';
      update(s => ({
        ...s,
        error: message,
      }));
    }
  }

  function disconnectWebSocket() {
    if (pollingInterval) {
      clearInterval(pollingInterval);
      pollingInterval = null;
    }
    if (updateWs) {
      updateWs.close();
      updateWs = null;
    }
  }

  return {
    subscribe,

    checkForUpdates: async () => {
      update(s => ({ ...s, isChecking: true, status: 'checking', error: null }));

      try {
        const response = await fetch('/api/update/check');
        if (!response.ok) {
          throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }

        const data: ApiUpdateStatus = await response.json();

        const hasUpdate = data.backend.updateAvailable || data.proxy.updateAvailable;

        update(s => ({
          ...s,
          isChecking: false,
          status: hasUpdate ? 'available' : 'idle',
          backend: data.backend,
          proxy: data.proxy,
          updateAvailable: hasUpdate,
        }));
      } catch (error) {
        const message = error instanceof Error ? error.message : 'Failed to check for updates';
        update(s => ({
          ...s,
          isChecking: false,
          status: 'error',
          error: message,
        }));
      }
    },

    startUpdate: async () => {
      // Get current state to check if proxy needs update
      const currentState = getState();

      // Reset flags
      backendUpdateSucceeded = false;
      proxyNeedsUpdate = currentState.proxy.updateAvailable;

      update(s => ({
        ...s,
        isUpdating: true,
        isReconnecting: false,
        status: 'updating',
        backendLogs: [],
        proxyLogs: [],
        error: null,
      }));

      try {
        // Connect to WebSocket for log streaming
        await connectWebSocket();

        // Start backend update first
        const backendResponse = await fetch('/api/update/backend', { method: 'POST' });
        if (!backendResponse.ok) {
          throw new Error(`Backend update failed: ${backendResponse.statusText}`);
        }

        // The logs will stream via WebSocket
        // When WebSocket closes (backend restart), it will:
        // 1. Poll until backend is back
        // 2. Reconnect WebSocket
        // 3. Trigger proxy update if needed

      } catch (error) {
        const message = error instanceof Error ? error.message : 'Update failed';
        update(s => ({
          ...s,
          isUpdating: false,
          isReconnecting: false,
          status: 'error',
          error: message,
        }));
        disconnectWebSocket();
      }
    },

    startProxyUpdate: async () => {
      try {
        const response = await fetch('/api/update/proxy', { method: 'POST' });
        if (!response.ok) {
          throw new Error(`Proxy update failed: ${response.statusText}`);
        }
        // Logs will stream via WebSocket
      } catch (error) {
        const message = error instanceof Error ? error.message : 'Proxy update failed';
        update(s => ({
          ...s,
          error: message,
        }));
      }
    },

    addBackendLog: (entry: LogEntry) => {
      update(s => ({
        ...s,
        backendLogs: [...s.backendLogs, entry],
      }));
    },

    addProxyLog: (entry: LogEntry) => {
      update(s => ({
        ...s,
        proxyLogs: [...s.proxyLogs, entry],
      }));
    },

    setUpdateAvailable: (available: boolean) => {
      update(s => ({
        ...s,
        updateAvailable: available,
        status: available ? 'available' : 'idle',
      }));
    },

    clearError: () => {
      update(s => ({ ...s, error: null }));
    },

    reset: () => {
      disconnectWebSocket();
      set(initialState);
    },

    disconnect: disconnectWebSocket,
  };
}

export const updateStore = createUpdateStore();

// Derived stores for convenience
export const updateAvailable = derived(updateStore, $s => $s.updateAvailable);
export const isUpdating = derived(updateStore, $s => $s.isUpdating);
export const isReconnecting = derived(updateStore, $s => $s.isReconnecting);
export const isChecking = derived(updateStore, $s => $s.isChecking);
export const updateError = derived(updateStore, $s => $s.error);
export const backendVersion = derived(updateStore, $s => $s.backend);
export const proxyVersion = derived(updateStore, $s => $s.proxy);
export const backendLogs = derived(updateStore, $s => $s.backendLogs);
export const proxyLogs = derived(updateStore, $s => $s.proxyLogs);
export const updateStatus = derived(updateStore, $s => $s.status);

// Legacy compatibility
export const versionInfo = derived(updateStore, $s => ({
  currentVersion: $s.backend.current,
  newVersion: $s.backend.latest,
}));
