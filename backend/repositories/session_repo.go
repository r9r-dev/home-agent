package repositories

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/ronan/home-agent/models"
)

// SQLiteSessionRepository implements SessionRepository using SQLite
type SQLiteSessionRepository struct {
	db *sql.DB
}

// NewSessionRepository creates a new SQLite session repository
func NewSessionRepository(db *sql.DB) SessionRepository {
	return &SQLiteSessionRepository{db: db}
}

// Create creates a new session with default model (haiku)
func (r *SQLiteSessionRepository) Create(sessionID string) (*models.Session, error) {
	return r.CreateWithModel(sessionID, "haiku")
}

// CreateWithModel creates a new session with specified model
func (r *SQLiteSessionRepository) CreateWithModel(sessionID, model string) (*models.Session, error) {
	now := time.Now()

	query := `
	INSERT INTO sessions (session_id, claude_session_id, title, model, created_at, last_activity)
	VALUES (?, '', '', ?, ?, ?)
	`

	result, err := r.db.Exec(query, sessionID, model, now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	log.Printf("Created new session: %s (ID: %d, model: %s)", sessionID, id, model)

	return &models.Session{
		ID:              int(id),
		SessionID:       sessionID,
		ClaudeSessionID: "",
		Title:           "",
		Model:           model,
		CreatedAt:       now,
		LastActivity:    now,
	}, nil
}

// Get retrieves a session by its session ID
func (r *SQLiteSessionRepository) Get(sessionID string) (*models.Session, error) {
	query := `
	SELECT id, session_id, COALESCE(claude_session_id, ''), title, COALESCE(model, 'haiku'), created_at, last_activity,
	       COALESCE(input_tokens, 0), COALESCE(output_tokens, 0), COALESCE(total_cost_usd, 0)
	FROM sessions
	WHERE session_id = ?
	`

	var session models.Session
	err := r.db.QueryRow(query, sessionID).Scan(
		&session.ID,
		&session.SessionID,
		&session.ClaudeSessionID,
		&session.Title,
		&session.Model,
		&session.CreatedAt,
		&session.LastActivity,
		&session.InputTokens,
		&session.OutputTokens,
		&session.TotalCostUSD,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &session, nil
}

// List retrieves all sessions ordered by last activity
func (r *SQLiteSessionRepository) List() ([]*models.Session, error) {
	query := `
	SELECT id, session_id, COALESCE(claude_session_id, ''), title, COALESCE(model, 'haiku'), created_at, last_activity,
	       COALESCE(input_tokens, 0), COALESCE(output_tokens, 0), COALESCE(total_cost_usd, 0)
	FROM sessions
	ORDER BY last_activity DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*models.Session
	for rows.Next() {
		var session models.Session
		err := rows.Scan(
			&session.ID,
			&session.SessionID,
			&session.ClaudeSessionID,
			&session.Title,
			&session.Model,
			&session.CreatedAt,
			&session.LastActivity,
			&session.InputTokens,
			&session.OutputTokens,
			&session.TotalCostUSD,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, &session)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sessions: %w", err)
	}

	return sessions, nil
}

// UpdateActivity updates the last activity timestamp for a session
func (r *SQLiteSessionRepository) UpdateActivity(sessionID string) error {
	query := `
	UPDATE sessions
	SET last_activity = ?
	WHERE session_id = ?
	`

	result, err := r.db.Exec(query, time.Now(), sessionID)
	if err != nil {
		return fmt.Errorf("failed to update session activity: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	return nil
}

// UpdateTitle updates the title of a session
func (r *SQLiteSessionRepository) UpdateTitle(sessionID, title string) error {
	query := `
	UPDATE sessions
	SET title = ?
	WHERE session_id = ?
	`

	result, err := r.db.Exec(query, title, sessionID)
	if err != nil {
		return fmt.Errorf("failed to update session title: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	return nil
}

// UpdateModel updates the model of a session
func (r *SQLiteSessionRepository) UpdateModel(sessionID, model string) error {
	query := `
	UPDATE sessions
	SET model = ?
	WHERE session_id = ?
	`

	result, err := r.db.Exec(query, model, sessionID)
	if err != nil {
		return fmt.Errorf("failed to update session model: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	log.Printf("Updated model for session %s: %s", sessionID, model)
	return nil
}

// UpdateClaudeSessionID updates the Claude CLI session ID
func (r *SQLiteSessionRepository) UpdateClaudeSessionID(sessionID, claudeSessionID string) error {
	query := `
	UPDATE sessions
	SET claude_session_id = ?
	WHERE session_id = ?
	`

	result, err := r.db.Exec(query, claudeSessionID, sessionID)
	if err != nil {
		return fmt.Errorf("failed to update claude session id: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	log.Printf("Updated Claude session ID for %s: %s", sessionID, claudeSessionID)
	return nil
}

// UpdateSessionID updates the session_id of a session and all related records
func (r *SQLiteSessionRepository) UpdateSessionID(oldSessionID, newSessionID string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update messages first (they reference session_id)
	_, err = tx.Exec("UPDATE messages SET session_id = ? WHERE session_id = ?", newSessionID, oldSessionID)
	if err != nil {
		return fmt.Errorf("failed to update messages session_id: %w", err)
	}

	// Update tool calls
	_, err = tx.Exec("UPDATE tool_calls SET session_id = ? WHERE session_id = ?", newSessionID, oldSessionID)
	if err != nil {
		return fmt.Errorf("failed to update tool_calls session_id: %w", err)
	}

	// Update session
	result, err := tx.Exec("UPDATE sessions SET session_id = ? WHERE session_id = ?", newSessionID, oldSessionID)
	if err != nil {
		return fmt.Errorf("failed to update session_id: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session not found: %s", oldSessionID)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Updated session ID: %s -> %s", oldSessionID, newSessionID)
	return nil
}

// Delete deletes a session and all related records
func (r *SQLiteSessionRepository) Delete(sessionID string) error {
	// Delete messages first (foreign key)
	_, err := r.db.Exec("DELETE FROM messages WHERE session_id = ?", sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete messages: %w", err)
	}

	// Delete tool calls
	_, err = r.db.Exec("DELETE FROM tool_calls WHERE session_id = ?", sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete tool calls: %w", err)
	}

	// Delete session
	result, err := r.db.Exec("DELETE FROM sessions WHERE session_id = ?", sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	log.Printf("Deleted session: %s", sessionID)
	return nil
}

// UpdateUsage updates the usage statistics for a session
func (r *SQLiteSessionRepository) UpdateUsage(sessionID string, inputTokens, outputTokens int, totalCostUSD float64) error {
	query := `
	UPDATE sessions
	SET input_tokens = ?, output_tokens = ?, total_cost_usd = ?
	WHERE session_id = ?
	`

	result, err := r.db.Exec(query, inputTokens, outputTokens, totalCostUSD, sessionID)
	if err != nil {
		return fmt.Errorf("failed to update session usage: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	return nil
}
