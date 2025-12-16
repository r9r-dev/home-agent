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
