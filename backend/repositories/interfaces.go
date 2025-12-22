package repositories

import (
	"github.com/ronan/home-agent/models"
)

// SessionRepository handles session persistence operations
type SessionRepository interface {
	Create(sessionID string) (*models.Session, error)
	CreateWithModel(sessionID, model string) (*models.Session, error)
	Get(sessionID string) (*models.Session, error)
	List() ([]*models.Session, error)
	UpdateActivity(sessionID string) error
	UpdateTitle(sessionID, title string) error
	UpdateModel(sessionID, model string) error
	UpdateClaudeSessionID(sessionID, claudeSessionID string) error
	UpdateSessionID(oldSessionID, newSessionID string) error
	Delete(sessionID string) error
}

// MessageRepository handles message persistence operations
type MessageRepository interface {
	Save(sessionID, role, content string) (*models.Message, error)
	GetBySession(sessionID string) ([]*models.Message, error)
}

// MemoryRepository handles memory entry persistence operations
type MemoryRepository interface {
	Create(id, title, content string) (*models.MemoryEntry, error)
	Get(id string) (*models.MemoryEntry, error)
	Update(id, title, content string, enabled bool) error
	Delete(id string) error
	List() ([]*models.MemoryEntry, error)
	GetEnabled() ([]*models.MemoryEntry, error)
}

// ToolCallRepository handles tool call persistence operations
type ToolCallRepository interface {
	Create(sessionID, toolUseID, toolName, input string) (*models.ToolCall, error)
	UpdateOutput(toolUseID, input, output, status string) error
	Get(toolUseID string) (*models.ToolCall, error)
	GetBySession(sessionID string) ([]*models.ToolCall, error)
}

// MachineRepository handles SSH machine persistence operations
type MachineRepository interface {
	Create(id, name, description, host string, port int, username, authType, encryptedAuthValue string) (*models.Machine, error)
	Get(id string) (*models.Machine, error)
	GetWithAuth(id string) (*models.Machine, error)
	List() ([]*models.Machine, error)
	Update(id, name, description, host string, port int, username, authType, encryptedAuthValue string) error
	UpdateStatus(id, status string) error
	Delete(id string) error
}

// SettingsRepository handles settings persistence operations
type SettingsRepository interface {
	Get(key string) (string, error)
	Set(key, value string) error
	GetAll() (map[string]string, error)
}

// SearchRepository handles full-text search operations
type SearchRepository interface {
	SearchMessages(query string, limit, offset int) ([]*models.SearchResult, int, error)
}
