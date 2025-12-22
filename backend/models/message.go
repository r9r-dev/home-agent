package models

import "time"

// Message represents a single message in a conversation
type Message struct {
	ID        int       `json:"id"`
	SessionID string    `json:"session_id"` // References Session.SessionID
	Role      string    `json:"role"`       // "user" or "assistant"
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}
