/**
 * Types for Claude Proxy SDK
 * Compatible with the existing Go backend protocol
 */

// Request from Home Agent backend
export interface ProxyRequest {
  type: "execute";
  prompt: string;
  session_id?: string;
  is_new_session?: boolean;
  model?: "haiku" | "sonnet" | "opus";
  custom_instructions?: string;
  thinking?: boolean;
}

// Response to Home Agent backend
export interface ProxyResponse {
  type: "chunk" | "thinking" | "session_id" | "done" | "error";
  content?: string;
  session_id?: string;
  error?: string;
}

// Configuration
export interface ProxyConfig {
  port: number;
  host: string;
  apiKey?: string;
}

// Hook event types
export type HookEventType =
  | "PreToolUse"
  | "PostToolUse"
  | "SessionStart"
  | "SessionEnd"
  | "Stop";

// Audit log entry
export interface AuditLogEntry {
  timestamp: Date;
  sessionId?: string;
  event: HookEventType | "execute" | "error";
  tool?: string;
  details?: Record<string, unknown>;
}
