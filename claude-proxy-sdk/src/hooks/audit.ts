/**
 * Audit logging hooks
 * Logs all Claude Agent activity for monitoring and debugging
 */

import type { AuditLogEntry } from "../types.js";

// In-memory log buffer (could be extended to file/database)
const auditBuffer: AuditLogEntry[] = [];
const MAX_BUFFER_SIZE = 1000;

/**
 * Log an audit entry
 */
export function auditLog(entry: AuditLogEntry): void {
  // Console output for development
  const logLine = formatLogEntry(entry);
  console.log(`[AUDIT] ${logLine}`);

  // Buffer for potential retrieval
  auditBuffer.push(entry);
  if (auditBuffer.length > MAX_BUFFER_SIZE) {
    auditBuffer.shift();
  }
}

/**
 * Format a log entry for output
 */
function formatLogEntry(entry: AuditLogEntry): string {
  const parts = [
    entry.timestamp.toISOString(),
    entry.event.toUpperCase(),
    entry.sessionId ? `session=${entry.sessionId.slice(0, 8)}` : "no-session",
  ];

  if (entry.tool) {
    parts.push(`tool=${entry.tool}`);
  }

  if (entry.details) {
    parts.push(JSON.stringify(entry.details));
  }

  return parts.join(" | ");
}

/**
 * Get recent audit entries
 */
export function getRecentAuditLogs(count = 100): AuditLogEntry[] {
  return auditBuffer.slice(-count);
}

/**
 * Clear audit buffer
 */
export function clearAuditBuffer(): void {
  auditBuffer.length = 0;
}
