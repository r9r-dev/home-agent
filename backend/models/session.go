package models

import "time"

// Session represents a conversation session with Claude Code
type Session struct {
	ID              int       `json:"id"`
	SessionID       string    `json:"session_id"`        // Internal session UUID
	ClaudeSessionID string    `json:"claude_session_id"` // Claude Code CLI session ID for --resume
	Title           string    `json:"title"`             // Auto-generated title from first message
	Model           string    `json:"model"`             // Claude model: haiku, sonnet, opus
	CreatedAt       time.Time `json:"created_at"`
	LastActivity    time.Time `json:"last_activity"`
}
