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

// Tool call information
export interface ToolCallInfo {
  tool_use_id: string;
  tool_name: string;
  input: Record<string, unknown>;
  parent_tool_use_id?: string | null;
}

// Response to Home Agent backend
export interface ProxyResponse {
  type: "chunk" | "thinking" | "session_id" | "done" | "error"
      | "tool_start" | "tool_progress" | "tool_result" | "tool_error"
      | "tool_input_delta";
  content?: string;
  session_id?: string;
  error?: string;
  // Tool-specific fields
  tool?: ToolCallInfo;
  elapsed_time_seconds?: number;
  tool_output?: string;
  is_error?: boolean;
  // Tool input streaming
  input_delta?: string;  // JSON delta for input streaming
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
