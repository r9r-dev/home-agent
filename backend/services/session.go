package services

import (
	"fmt"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/ronan/home-agent/models"
)

// SessionManager manages conversation sessions
type SessionManager struct {
	db       *models.DB
	sessions sync.Map // Map of web session IDs to Claude session IDs
	mu       sync.RWMutex
}

// NewSessionManager creates a new SessionManager instance
func NewSessionManager(db *models.DB) *SessionManager {
	log.Println("Initializing SessionManager")
	return &SessionManager{
		db: db,
	}
}

// CreateSession creates a new session with a generated UUID
func (sm *SessionManager) CreateSession() (string, error) {
	sessionID := uuid.New().String()

	session, err := sm.db.CreateSession(sessionID)
	if err != nil {
		return "", fmt.Errorf("failed to create session in database: %w", err)
	}

	log.Printf("SessionManager: Created new session %s", session.SessionID)
	return session.SessionID, nil
}

// GetSession retrieves a session by ID
func (sm *SessionManager) GetSession(sessionID string) (*models.Session, error) {
	session, err := sm.db.GetSession(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session from database: %w", err)
	}

	if session == nil {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	return session, nil
}

// UpdateSessionActivity updates the last activity timestamp for a session
func (sm *SessionManager) UpdateSessionActivity(sessionID string) error {
	if err := sm.db.UpdateSessionActivity(sessionID); err != nil {
		return fmt.Errorf("failed to update session activity: %w", err)
	}
	return nil
}

// SaveMessage saves a message to the database
func (sm *SessionManager) SaveMessage(sessionID, role, content string) error {
	// Validate role
	if role != "user" && role != "assistant" {
		return fmt.Errorf("invalid role: %s (must be 'user' or 'assistant')", role)
	}

	_, err := sm.db.SaveMessage(sessionID, role, content)
	if err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}

	// Update session activity
	if err := sm.UpdateSessionActivity(sessionID); err != nil {
		log.Printf("Warning: failed to update session activity: %v", err)
	}

	return nil
}

// GetMessages retrieves all messages for a session
func (sm *SessionManager) GetMessages(sessionID string) ([]*models.Message, error) {
	messages, err := sm.db.GetMessages(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	return messages, nil
}

// SessionExists checks if a session exists in the database
func (sm *SessionManager) SessionExists(sessionID string) bool {
	session, err := sm.db.GetSession(sessionID)
	if err != nil {
		log.Printf("Error checking session existence: %v", err)
		return false
	}
	return session != nil
}

// MapWebSessionToClaude maps a web session ID to a Claude session ID
// This is useful when the frontend wants to maintain its own session IDs
func (sm *SessionManager) MapWebSessionToClaude(webSessionID, claudeSessionID string) {
	sm.sessions.Store(webSessionID, claudeSessionID)
	log.Printf("Mapped web session %s to Claude session %s", webSessionID, claudeSessionID)
}

// GetClaudeSessionID gets the Claude session ID for a web session ID
func (sm *SessionManager) GetClaudeSessionID(webSessionID string) (string, bool) {
	value, ok := sm.sessions.Load(webSessionID)
	if !ok {
		return "", false
	}
	claudeSessionID, ok := value.(string)
	return claudeSessionID, ok
}

// UnmapWebSession removes the mapping for a web session
func (sm *SessionManager) UnmapWebSession(webSessionID string) {
	sm.sessions.Delete(webSessionID)
	log.Printf("Unmapped web session %s", webSessionID)
}

// GetOrCreateSession gets an existing session or creates a new one if the ID is empty
func (sm *SessionManager) GetOrCreateSession(sessionID string) (string, bool, error) {
	// If no session ID provided, create a new one
	if sessionID == "" {
		newSessionID, err := sm.CreateSession()
		if err != nil {
			return "", false, fmt.Errorf("failed to create new session: %w", err)
		}
		return newSessionID, true, nil
	}

	// Check if session exists
	exists := sm.SessionExists(sessionID)
	if !exists {
		return "", false, fmt.Errorf("session does not exist: %s", sessionID)
	}

	return sessionID, false, nil
}

// ListSessions returns all sessions ordered by last activity
func (sm *SessionManager) ListSessions() ([]*models.Session, error) {
	sessions, err := sm.db.ListSessions()
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	return sessions, nil
}

// UpdateSessionTitle updates the title of a session
func (sm *SessionManager) UpdateSessionTitle(sessionID, title string) error {
	if err := sm.db.UpdateSessionTitle(sessionID, title); err != nil {
		return fmt.Errorf("failed to update session title: %w", err)
	}
	return nil
}

// DeleteSession deletes a session and all its messages
func (sm *SessionManager) DeleteSession(sessionID string) error {
	if err := sm.db.DeleteSession(sessionID); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

// GenerateTitle generates a title from the first user message (max 50 chars)
func GenerateTitle(content string) string {
	// Remove newlines and extra spaces
	title := content
	if len(title) > 50 {
		title = title[:47] + "..."
	}
	return title
}
