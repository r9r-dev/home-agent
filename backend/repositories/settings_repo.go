package repositories

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

// SQLiteSettingsRepository implements SettingsRepository using SQLite
type SQLiteSettingsRepository struct {
	db *sql.DB
}

// NewSettingsRepository creates a new SQLite settings repository
func NewSettingsRepository(db *sql.DB) SettingsRepository {
	return &SQLiteSettingsRepository{db: db}
}

// Get retrieves a setting value by key
func (r *SQLiteSettingsRepository) Get(key string) (string, error) {
	query := `SELECT value FROM settings WHERE key = ?`

	var value string
	err := r.db.QueryRow(query, key).Scan(&value)

	if err == sql.ErrNoRows {
		return "", nil
	}

	if err != nil {
		return "", fmt.Errorf("failed to get setting: %w", err)
	}

	return value, nil
}

// Set creates or updates a setting
func (r *SQLiteSettingsRepository) Set(key, value string) error {
	query := `
	INSERT INTO settings (key, value, updated_at)
	VALUES (?, ?, ?)
	ON CONFLICT(key) DO UPDATE SET value = ?, updated_at = ?
	`

	now := time.Now()
	_, err := r.db.Exec(query, key, value, now, value, now)
	if err != nil {
		return fmt.Errorf("failed to set setting: %w", err)
	}

	log.Printf("Setting updated: %s", key)
	return nil
}

// GetAll retrieves all settings as a map
func (r *SQLiteSettingsRepository) GetAll() (map[string]string, error) {
	query := `SELECT key, value FROM settings`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get settings: %w", err)
	}
	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("failed to scan setting: %w", err)
		}
		settings[key] = value
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating settings: %w", err)
	}

	return settings, nil
}
