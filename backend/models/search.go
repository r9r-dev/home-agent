package models

import "time"

// SearchResult represents a search result with snippet
type SearchResult struct {
	MessageID    int       `json:"message_id"`
	SessionID    string    `json:"session_id"`
	Role         string    `json:"role"`
	Snippet      string    `json:"snippet"`
	Timestamp    time.Time `json:"timestamp"`
	SessionTitle string    `json:"session_title"`
}
