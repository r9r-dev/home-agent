package repositories

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/ronan/home-agent/models"
)

// SQLiteMessageRepository implements MessageRepository using SQLite
type SQLiteMessageRepository struct {
	db *sql.DB
}

// NewMessageRepository creates a new SQLite message repository
func NewMessageRepository(db *sql.DB) MessageRepository {
	return &SQLiteMessageRepository{db: db}
}

// Save saves a message to the database
func (r *SQLiteMessageRepository) Save(sessionID, role, content string) (*models.Message, error) {
	now := time.Now()

	query := `
	INSERT INTO messages (session_id, role, content, created_at)
	VALUES (?, ?, ?, ?)
	`

	result, err := r.db.Exec(query, sessionID, role, content, now)
	if err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	log.Printf("Saved %s message for session %s (ID: %d)", role, sessionID, id)

	return &models.Message{
		ID:        int(id),
		SessionID: sessionID,
		Role:      role,
		Content:   content,
		CreatedAt: now,
	}, nil
}

// GetBySession retrieves all messages for a session, ordered by creation time
func (r *SQLiteMessageRepository) GetBySession(sessionID string) ([]*models.Message, error) {
	query := `
	SELECT id, session_id, role, content, created_at
	FROM messages
	WHERE session_id = ?
	ORDER BY created_at ASC
	`

	rows, err := r.db.Query(query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	defer rows.Close()

	var messages []*models.Message
	for rows.Next() {
		var msg models.Message
		err := rows.Scan(
			&msg.ID,
			&msg.SessionID,
			&msg.Role,
			&msg.Content,
			&msg.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, &msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating messages: %w", err)
	}

	return messages, nil
}
