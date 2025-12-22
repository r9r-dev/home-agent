package models

import "time"

// ToolCall represents a tool call made by Claude
type ToolCall struct {
	ID          int        `json:"id"`
	SessionID   string     `json:"session_id"`
	ToolUseID   string     `json:"tool_use_id"`
	ToolName    string     `json:"tool_name"`
	Input       string     `json:"input"`  // JSON string
	Output      string     `json:"output"` // JSON string or text
	Status      string     `json:"status"` // "running", "success", "error"
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}
