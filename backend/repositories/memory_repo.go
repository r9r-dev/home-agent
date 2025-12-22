package repositories

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/ronan/home-agent/models"
)

// SQLiteMemoryRepository implements MemoryRepository using SQLite
type SQLiteMemoryRepository struct {
	db *sql.DB
}

// NewMemoryRepository creates a new SQLite memory repository
func NewMemoryRepository(db *sql.DB) MemoryRepository {
	return &SQLiteMemoryRepository{db: db}
}

// Create creates a new memory entry
func (r *SQLiteMemoryRepository) Create(id, title, content string) (*models.MemoryEntry, error) {
	now := time.Now()

	query := `
	INSERT INTO memory (id, title, content, enabled, created_at, updated_at)
	VALUES (?, ?, ?, 1, ?, ?)
	`

	_, err := r.db.Exec(query, id, title, content, now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to create memory entry: %w", err)
	}

	log.Printf("Created memory entry: %s", id)

	return &models.MemoryEntry{
		ID:        id,
		Title:     title,
		Content:   content,
		Enabled:   true,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Get retrieves a memory entry by ID
func (r *SQLiteMemoryRepository) Get(id string) (*models.MemoryEntry, error) {
	query := `
	SELECT id, title, content, enabled, created_at, updated_at
	FROM memory
	WHERE id = ?
	`

	var entry models.MemoryEntry
	var enabled int
	err := r.db.QueryRow(query, id).Scan(
		&entry.ID,
		&entry.Title,
		&entry.Content,
		&enabled,
		&entry.CreatedAt,
		&entry.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get memory entry: %w", err)
	}

	entry.Enabled = enabled == 1
	return &entry, nil
}

// Update updates an existing memory entry
func (r *SQLiteMemoryRepository) Update(id, title, content string, enabled bool) error {
	now := time.Now()
	enabledInt := 0
	if enabled {
		enabledInt = 1
	}

	query := `
	UPDATE memory
	SET title = ?, content = ?, enabled = ?, updated_at = ?
	WHERE id = ?
	`

	result, err := r.db.Exec(query, title, content, enabledInt, now, id)
	if err != nil {
		return fmt.Errorf("failed to update memory entry: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("memory entry not found: %s", id)
	}

	log.Printf("Updated memory entry: %s", id)
	return nil
}

// Delete deletes a memory entry
func (r *SQLiteMemoryRepository) Delete(id string) error {
	result, err := r.db.Exec("DELETE FROM memory WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete memory entry: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("memory entry not found: %s", id)
	}

	log.Printf("Deleted memory entry: %s", id)
	return nil
}

// List retrieves all memory entries
func (r *SQLiteMemoryRepository) List() ([]*models.MemoryEntry, error) {
	query := `
	SELECT id, title, content, enabled, created_at, updated_at
	FROM memory
	ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list memory entries: %w", err)
	}
	defer rows.Close()

	var entries []*models.MemoryEntry
	for rows.Next() {
		var entry models.MemoryEntry
		var enabled int
		err := rows.Scan(
			&entry.ID,
			&entry.Title,
			&entry.Content,
			&enabled,
			&entry.CreatedAt,
			&entry.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan memory entry: %w", err)
		}
		entry.Enabled = enabled == 1
		entries = append(entries, &entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating memory entries: %w", err)
	}

	return entries, nil
}

// GetEnabled retrieves only enabled memory entries
func (r *SQLiteMemoryRepository) GetEnabled() ([]*models.MemoryEntry, error) {
	query := `
	SELECT id, title, content, enabled, created_at, updated_at
	FROM memory
	WHERE enabled = 1
	ORDER BY created_at ASC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get enabled memory entries: %w", err)
	}
	defer rows.Close()

	var entries []*models.MemoryEntry
	for rows.Next() {
		var entry models.MemoryEntry
		var enabled int
		err := rows.Scan(
			&entry.ID,
			&entry.Title,
			&entry.Content,
			&enabled,
			&entry.CreatedAt,
			&entry.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan memory entry: %w", err)
		}
		entry.Enabled = enabled == 1
		entries = append(entries, &entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating memory entries: %w", err)
	}

	return entries, nil
}
