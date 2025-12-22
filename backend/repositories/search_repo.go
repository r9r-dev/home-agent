package repositories

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/ronan/home-agent/models"
)

// SQLiteSearchRepository implements SearchRepository using SQLite FTS5
type SQLiteSearchRepository struct {
	db *sql.DB
}

// NewSearchRepository creates a new SQLite search repository
func NewSearchRepository(db *sql.DB) SearchRepository {
	return &SQLiteSearchRepository{db: db}
}

// SearchMessages performs full-text search on messages
func (r *SQLiteSearchRepository) SearchMessages(query string, limit, offset int) ([]*models.SearchResult, int, error) {
	// Sanitize query for FTS5 (wrap in quotes to handle special characters)
	safeQuery := `"` + strings.ReplaceAll(query, `"`, `""`) + `"`

	// Count total matches
	var total int
	countQuery := `
		SELECT COUNT(*) FROM messages_fts
		WHERE messages_fts MATCH ?
	`
	if err := r.db.QueryRow(countQuery, safeQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count search results: %w", err)
	}

	// Get results with snippets
	searchQuery := `
		SELECT
			m.id,
			m.session_id,
			m.role,
			snippet(messages_fts, 2, '<mark>', '</mark>', '...', 32) as snippet,
			m.created_at,
			COALESCE(s.title, '') as session_title
		FROM messages_fts
		JOIN messages m ON messages_fts.rowid = m.id
		JOIN sessions s ON m.session_id = s.session_id
		WHERE messages_fts MATCH ?
		ORDER BY rank
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.Query(searchQuery, safeQuery, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search messages: %w", err)
	}
	defer rows.Close()

	var results []*models.SearchResult
	for rows.Next() {
		var result models.SearchResult
		if err := rows.Scan(&result.MessageID, &result.SessionID, &result.Role, &result.Snippet, &result.Timestamp, &result.SessionTitle); err != nil {
			return nil, 0, fmt.Errorf("failed to scan search result: %w", err)
		}
		results = append(results, &result)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating search results: %w", err)
	}

	return results, total, nil
}
