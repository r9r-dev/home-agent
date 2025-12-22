package services

import (
	"fmt"
	"log"
	"sync"

	"github.com/ronan/home-agent/models"
	"github.com/ronan/home-agent/repositories"
)

// SessionManager manages conversation sessions
type SessionManager struct {
	sessions    repositories.SessionRepository
	messages    repositories.MessageRepository
	sessionsMap sync.Map // Map of web session IDs to Claude session IDs
	mu          sync.RWMutex
}

// NewSessionManager creates a new SessionManager instance
func NewSessionManager(sessions repositories.SessionRepository, messages repositories.MessageRepository) *SessionManager {
	return &SessionManager{
		sessions: sessions,
		messages: messages,
	}
}

// CreateSessionWithID creates a new session with a specific ID (from SDK) and model
func (sm *SessionManager) CreateSessionWithID(sessionID, model string) (*models.Session, error) {
	session, err := sm.sessions.CreateWithModel(sessionID, model)
	if err != nil {
		return nil, fmt.Errorf("failed to create session in database: %w", err)
	}

	log.Printf("SessionManager: Created new session %s with model %s", session.SessionID, model)
	return session, nil
}

// GetSession retrieves a session by ID
func (sm *SessionManager) GetSession(sessionID string) (*models.Session, error) {
	session, err := sm.sessions.Get(sessionID)
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
	if err := sm.sessions.UpdateActivity(sessionID); err != nil {
		return fmt.Errorf("failed to update session activity: %w", err)
	}
	return nil
}

// SaveMessage saves a message to the database
func (sm *SessionManager) SaveMessage(sessionID, role, content string) error {
	// Validate role
	if role != "user" && role != "assistant" && role != "thinking" {
		return fmt.Errorf("invalid role: %s (must be 'user', 'assistant', or 'thinking')", role)
	}

	_, err := sm.messages.Save(sessionID, role, content)
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
	messages, err := sm.messages.GetBySession(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	return messages, nil
}

// SessionExists checks if a session exists in the database
func (sm *SessionManager) SessionExists(sessionID string) bool {
	session, err := sm.sessions.Get(sessionID)
	if err != nil {
		log.Printf("Error checking session existence: %v", err)
		return false
	}
	return session != nil
}

// UpdateSessionID updates the session ID when SDK returns a new one after resume
func (sm *SessionManager) UpdateSessionID(oldSessionID, newSessionID string) error {
	if err := sm.sessions.UpdateSessionID(oldSessionID, newSessionID); err != nil {
		return fmt.Errorf("failed to update session id: %w", err)
	}
	return nil
}

// ListSessions returns all sessions ordered by last activity
func (sm *SessionManager) ListSessions() ([]*models.Session, error) {
	sessions, err := sm.sessions.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	return sessions, nil
}

// UpdateSessionTitle updates the title of a session
func (sm *SessionManager) UpdateSessionTitle(sessionID, title string) error {
	if err := sm.sessions.UpdateTitle(sessionID, title); err != nil {
		return fmt.Errorf("failed to update session title: %w", err)
	}
	return nil
}

// DeleteSession deletes a session and all its messages
func (sm *SessionManager) DeleteSession(sessionID string) error {
	if err := sm.sessions.Delete(sessionID); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

// UpdateSessionModel updates the model of a session
func (sm *SessionManager) UpdateSessionModel(sessionID, model string) error {
	if err := sm.sessions.UpdateModel(sessionID, model); err != nil {
		return fmt.Errorf("failed to update session model: %w", err)
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
