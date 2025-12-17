/**
 * API Service for REST endpoints
 */

export type ClaudeModel = 'haiku' | 'sonnet' | 'opus';

export interface Session {
  id: number;
  session_id: string;
  title: string;
  model: ClaudeModel;
  created_at: string;
  last_activity: string;
}

export interface Message {
  id: number;
  session_id: string;
  role: 'user' | 'assistant';
  content: string;
  created_at: string;
}

export interface UploadedFile {
  id: string;
  filename: string;
  path: string;
  type: 'image' | 'file';
  size: number;
  mime_type: string;
}

export interface MemoryEntry {
  id: string;
  title: string;
  content: string;
  enabled: boolean;
  created_at: string;
  updated_at: string;
}

export interface MemoryExport {
  version: string;
  entries: MemoryEntry[];
}

export interface MemoryImportResult {
  imported: number;
  errors: string[];
}

const API_BASE = '/api';

/**
 * Fetch all sessions
 */
export async function fetchSessions(): Promise<Session[]> {
  const response = await fetch(`${API_BASE}/sessions`);
  if (!response.ok) {
    throw new Error('Failed to fetch sessions');
  }
  const data = await response.json();
  return data || [];
}

/**
 * Fetch messages for a session
 */
export async function fetchMessages(sessionId: string): Promise<Message[]> {
  const response = await fetch(`${API_BASE}/sessions/${sessionId}/messages`);
  if (!response.ok) {
    throw new Error('Failed to fetch messages');
  }
  const data = await response.json();
  return data || [];
}

/**
 * Fetch a single session by ID
 */
export async function fetchSession(sessionId: string): Promise<Session> {
  const response = await fetch(`${API_BASE}/sessions/${sessionId}`);
  if (!response.ok) {
    throw new Error('Failed to fetch session');
  }
  return response.json();
}

/**
 * Delete a session
 */
export async function deleteSession(sessionId: string): Promise<void> {
  const response = await fetch(`${API_BASE}/sessions/${sessionId}`, {
    method: 'DELETE',
  });
  if (!response.ok) {
    throw new Error('Failed to delete session');
  }
}

/**
 * Update session model
 */
export async function updateSessionModel(sessionId: string, model: ClaudeModel): Promise<void> {
  const response = await fetch(`${API_BASE}/sessions/${sessionId}/model`, {
    method: 'PATCH',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ model }),
  });
  if (!response.ok) {
    throw new Error('Failed to update session model');
  }
}

/**
 * Upload a file
 */
export async function uploadFile(file: File, sessionId?: string): Promise<UploadedFile> {
  const formData = new FormData();
  formData.append('file', file);
  if (sessionId) {
    formData.append('session_id', sessionId);
  }

  const response = await fetch(`${API_BASE}/upload`, {
    method: 'POST',
    body: formData,
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: 'Upload failed' }));
    throw new Error(error.error || 'Failed to upload file');
  }

  return response.json();
}

/**
 * Delete an uploaded file
 */
export async function deleteUploadedFile(fileId: string, sessionId?: string): Promise<void> {
  const url = sessionId
    ? `${API_BASE}/uploads/${fileId}?session_id=${sessionId}`
    : `${API_BASE}/uploads/${fileId}`;

  const response = await fetch(url, {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error('Failed to delete file');
  }
}

/**
 * Fetch all settings
 */
export async function fetchSettings(): Promise<Record<string, string>> {
  const response = await fetch(`${API_BASE}/settings`);
  if (!response.ok) {
    throw new Error('Failed to fetch settings');
  }
  const data = await response.json();
  return data || {};
}

/**
 * Update a setting
 */
export async function updateSetting(key: string, value: string): Promise<void> {
  const response = await fetch(`${API_BASE}/settings/${key}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ value }),
  });
  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: 'Update failed' }));
    throw new Error(error.error || 'Failed to update setting');
  }
}

/**
 * Fetch the base system prompt
 */
export async function fetchSystemPrompt(): Promise<string> {
  const response = await fetch(`${API_BASE}/system-prompt`);
  if (!response.ok) {
    throw new Error('Failed to fetch system prompt');
  }
  const data = await response.json();
  return data.prompt || '';
}

// Memory API functions

/**
 * Fetch all memory entries
 */
export async function fetchMemoryEntries(): Promise<MemoryEntry[]> {
  const response = await fetch(`${API_BASE}/memory`);
  if (!response.ok) {
    throw new Error('Failed to fetch memory entries');
  }
  const data = await response.json();
  return data || [];
}

/**
 * Create a new memory entry
 */
export async function createMemoryEntry(title: string, content: string): Promise<MemoryEntry> {
  const response = await fetch(`${API_BASE}/memory`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ title, content }),
  });
  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: 'Creation failed' }));
    throw new Error(error.error || 'Failed to create memory entry');
  }
  return response.json();
}

/**
 * Update a memory entry
 */
export async function updateMemoryEntry(
  id: string,
  data: { title?: string; content?: string; enabled?: boolean }
): Promise<MemoryEntry> {
  const response = await fetch(`${API_BASE}/memory/${id}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(data),
  });
  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: 'Update failed' }));
    throw new Error(error.error || 'Failed to update memory entry');
  }
  return response.json();
}

/**
 * Delete a memory entry
 */
export async function deleteMemoryEntry(id: string): Promise<void> {
  const response = await fetch(`${API_BASE}/memory/${id}`, {
    method: 'DELETE',
  });
  if (!response.ok) {
    throw new Error('Failed to delete memory entry');
  }
}

/**
 * Export all memory entries
 */
export async function exportMemory(): Promise<MemoryExport> {
  const response = await fetch(`${API_BASE}/memory/export`);
  if (!response.ok) {
    throw new Error('Failed to export memory');
  }
  return response.json();
}

/**
 * Import memory entries
 */
export async function importMemory(
  entries: Array<{ title: string; content: string; enabled: boolean }>
): Promise<MemoryImportResult> {
  const response = await fetch(`${API_BASE}/memory/import`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ entries }),
  });
  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: 'Import failed' }));
    throw new Error(error.error || 'Failed to import memory');
  }
  return response.json();
}
