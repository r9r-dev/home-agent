package repositories

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/ronan/home-agent/models"
)

// SQLiteToolCallRepository implements ToolCallRepository using SQLite
type SQLiteToolCallRepository struct {
	db *sql.DB
}

// NewToolCallRepository creates a new SQLite tool call repository
func NewToolCallRepository(db *sql.DB) ToolCallRepository {
	return &SQLiteToolCallRepository{db: db}
}

// Create creates a new tool call record
func (r *SQLiteToolCallRepository) Create(sessionID, toolUseID, toolName, input string) (*models.ToolCall, error) {
	now := time.Now()

	query := `
	INSERT INTO tool_calls (session_id, tool_use_id, tool_name, input, output, status, created_at)
	VALUES (?, ?, ?, ?, '', 'running', ?)
	`

	result, err := r.db.Exec(query, sessionID, toolUseID, toolName, input, now)
	if err != nil {
		return nil, fmt.Errorf("failed to create tool call: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	log.Printf("Created tool call: %s (%s) for session %s", toolName, toolUseID, sessionID)

	return &models.ToolCall{
		ID:        int(id),
		SessionID: sessionID,
		ToolUseID: toolUseID,
		ToolName:  toolName,
		Input:     input,
		Output:    "",
		Status:    "running",
		CreatedAt: now,
	}, nil
}

// UpdateOutput updates a tool call with its input, output, and status
func (r *SQLiteToolCallRepository) UpdateOutput(toolUseID, input, output, status string) error {
	now := time.Now()

	query := `
	UPDATE tool_calls
	SET input = ?, output = ?, status = ?, completed_at = ?
	WHERE tool_use_id = ?
	`

	result, err := r.db.Exec(query, input, output, status, now, toolUseID)
	if err != nil {
		return fmt.Errorf("failed to update tool call: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("tool call not found: %s", toolUseID)
	}

	log.Printf("Updated tool call %s: status=%s", toolUseID, status)
	return nil
}

// Get retrieves a tool call by its tool_use_id
func (r *SQLiteToolCallRepository) Get(toolUseID string) (*models.ToolCall, error) {
	query := `
	SELECT id, session_id, tool_use_id, tool_name, input, output, status, created_at, completed_at
	FROM tool_calls
	WHERE tool_use_id = ?
	`

	var tc models.ToolCall
	var completedAt sql.NullTime
	err := r.db.QueryRow(query, toolUseID).Scan(
		&tc.ID,
		&tc.SessionID,
		&tc.ToolUseID,
		&tc.ToolName,
		&tc.Input,
		&tc.Output,
		&tc.Status,
		&tc.CreatedAt,
		&completedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get tool call: %w", err)
	}

	if completedAt.Valid {
		tc.CompletedAt = &completedAt.Time
	}

	return &tc, nil
}

// GetBySession retrieves all tool calls for a session
func (r *SQLiteToolCallRepository) GetBySession(sessionID string) ([]*models.ToolCall, error) {
	query := `
	SELECT id, session_id, tool_use_id, tool_name, input, output, status, created_at, completed_at
	FROM tool_calls
	WHERE session_id = ?
	ORDER BY created_at ASC
	`

	rows, err := r.db.Query(query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tool calls: %w", err)
	}
	defer rows.Close()

	var toolCalls []*models.ToolCall
	for rows.Next() {
		var tc models.ToolCall
		var completedAt sql.NullTime
		err := rows.Scan(
			&tc.ID,
			&tc.SessionID,
			&tc.ToolUseID,
			&tc.ToolName,
			&tc.Input,
			&tc.Output,
			&tc.Status,
			&tc.CreatedAt,
			&completedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tool call: %w", err)
		}
		if completedAt.Valid {
			tc.CompletedAt = &completedAt.Time
		}
		toolCalls = append(toolCalls, &tc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tool calls: %w", err)
	}

	return toolCalls, nil
}
