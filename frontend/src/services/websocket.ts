/**
 * WebSocket Service for real-time communication with the backend
 * Handles connection, reconnection, and message exchange
 */

export type MessageHandler = (data: any) => void;
export type ErrorHandler = (error: Event) => void;
export type CloseHandler = (event: CloseEvent) => void;
export type OpenHandler = () => void;

interface WebSocketConfig {
  url: string;
  reconnectInterval?: number;
  maxReconnectInterval?: number;
  reconnectDecay?: number;
  maxReconnectAttempts?: number;
}

export class WebSocketService {
  private ws: WebSocket | null = null;
  private config: Required<WebSocketConfig>;
  private reconnectAttempts = 0;
  private reconnectTimeout: number | null = null;
  private shouldReconnect = true;
  private currentReconnectInterval: number;

  // Event handlers
  private messageHandlers: MessageHandler[] = [];
  private errorHandlers: ErrorHandler[] = [];
  private closeHandlers: CloseHandler[] = [];
  private openHandlers: OpenHandler[] = [];

  constructor(config: WebSocketConfig) {
    this.config = {
      url: config.url,
      reconnectInterval: config.reconnectInterval ?? 1000,
      maxReconnectInterval: config.maxReconnectInterval ?? 30000,
      reconnectDecay: config.reconnectDecay ?? 1.5,
      maxReconnectAttempts: config.maxReconnectAttempts ?? Infinity,
    };
    this.currentReconnectInterval = this.config.reconnectInterval;
  }

  /**
   * Connect to the WebSocket server
   */
  connect(): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      console.log('[WebSocket] Already connected');
      return;
    }

    try {
      console.log(`[WebSocket] Connecting to ${this.config.url}...`);
      this.ws = new WebSocket(this.config.url);

      this.ws.onopen = () => {
        console.log('[WebSocket] Connected');
        this.reconnectAttempts = 0;
        this.currentReconnectInterval = this.config.reconnectInterval;
        this.openHandlers.forEach((handler) => handler());
      };

      this.ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          this.messageHandlers.forEach((handler) => handler(data));
        } catch (error) {
          console.error('[WebSocket] Failed to parse message:', error);
        }
      };

      this.ws.onerror = (error) => {
        console.error('[WebSocket] Error:', error);
        this.errorHandlers.forEach((handler) => handler(error));
      };

      this.ws.onclose = (event) => {
        console.log(`[WebSocket] Disconnected (code: ${event.code})`);
        this.closeHandlers.forEach((handler) => handler(event));

        if (this.shouldReconnect && this.reconnectAttempts < this.config.maxReconnectAttempts) {
          this.scheduleReconnect();
        }
      };
    } catch (error) {
      console.error('[WebSocket] Connection failed:', error);
      if (this.shouldReconnect && this.reconnectAttempts < this.config.maxReconnectAttempts) {
        this.scheduleReconnect();
      }
    }
  }

  /**
   * Schedule reconnection with exponential backoff
   */
  private scheduleReconnect(): void {
    if (this.reconnectTimeout !== null) {
      return;
    }

    this.reconnectAttempts++;
    console.log(
      `[WebSocket] Reconnecting in ${this.currentReconnectInterval}ms (attempt ${this.reconnectAttempts})`
    );

    this.reconnectTimeout = window.setTimeout(() => {
      this.reconnectTimeout = null;
      this.connect();
    }, this.currentReconnectInterval);

    // Exponential backoff
    this.currentReconnectInterval = Math.min(
      this.currentReconnectInterval * this.config.reconnectDecay,
      this.config.maxReconnectInterval
    );
  }

  /**
   * Disconnect from the WebSocket server
   */
  disconnect(): void {
    this.shouldReconnect = false;

    if (this.reconnectTimeout !== null) {
      clearTimeout(this.reconnectTimeout);
      this.reconnectTimeout = null;
    }

    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }

    console.log('[WebSocket] Disconnected by user');
  }

  /**
   * Send a message to the server
   */
  sendMessage(content: string, sessionId?: string): void {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      console.error('[WebSocket] Cannot send message: not connected');
      throw new Error('WebSocket is not connected');
    }

    const message: { type: string; content: string; sessionId?: string } = {
      type: 'message',
      content,
    };

    if (sessionId) {
      message.sessionId = sessionId;
    }

    this.ws.send(JSON.stringify(message));
    console.log('[WebSocket] Message sent (sessionId:', sessionId || 'none', ')');
  }

  /**
   * Register a message handler
   */
  onMessage(handler: MessageHandler): () => void {
    this.messageHandlers.push(handler);
    return () => {
      this.messageHandlers = this.messageHandlers.filter((h) => h !== handler);
    };
  }

  /**
   * Register an error handler
   */
  onError(handler: ErrorHandler): () => void {
    this.errorHandlers.push(handler);
    return () => {
      this.errorHandlers = this.errorHandlers.filter((h) => h !== handler);
    };
  }

  /**
   * Register a close handler
   */
  onClose(handler: CloseHandler): () => void {
    this.closeHandlers.push(handler);
    return () => {
      this.closeHandlers = this.closeHandlers.filter((h) => h !== handler);
    };
  }

  /**
   * Register an open handler
   */
  onOpen(handler: OpenHandler): () => void {
    this.openHandlers.push(handler);
    return () => {
      this.openHandlers = this.openHandlers.filter((h) => h !== handler);
    };
  }

  /**
   * Check if connected
   */
  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }

  /**
   * Get connection state
   */
  getReadyState(): number | null {
    return this.ws?.readyState ?? null;
  }
}

// Build WebSocket URL from current location
function getWebSocketUrl(): string {
  // Use env variable if set (for development)
  if (import.meta.env.VITE_WS_URL) {
    return import.meta.env.VITE_WS_URL;
  }

  // Auto-detect from current location
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const host = window.location.host;
  return `${protocol}//${host}/ws`;
}

// Singleton instance
export const websocketService = new WebSocketService({ url: getWebSocketUrl() });
