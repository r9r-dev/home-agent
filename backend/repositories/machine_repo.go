package repositories

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/ronan/home-agent/models"
)

// SQLiteMachineRepository implements MachineRepository using SQLite
type SQLiteMachineRepository struct {
	db *sql.DB
}

// NewMachineRepository creates a new SQLite machine repository
func NewMachineRepository(db *sql.DB) MachineRepository {
	return &SQLiteMachineRepository{db: db}
}

// Create creates a new machine entry
func (r *SQLiteMachineRepository) Create(id, name, description, host string, port int, username, authType, encryptedAuthValue string) (*models.Machine, error) {
	now := time.Now()

	query := `
	INSERT INTO machines (id, name, description, host, port, username, auth_type, auth_value, status, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'untested', ?, ?)
	`

	_, err := r.db.Exec(query, id, name, description, host, port, username, authType, encryptedAuthValue, now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to create machine: %w", err)
	}

	log.Printf("Created machine: %s (%s)", name, id)

	return &models.Machine{
		ID:          id,
		Name:        name,
		Description: description,
		Host:        host,
		Port:        port,
		Username:    username,
		AuthType:    authType,
		AuthValue:   encryptedAuthValue,
		Status:      "untested",
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// Get retrieves a machine by ID (without auth_value)
func (r *SQLiteMachineRepository) Get(id string) (*models.Machine, error) {
	query := `
	SELECT id, name, description, host, port, username, auth_type, status, created_at, updated_at
	FROM machines
	WHERE id = ?
	`

	var machine models.Machine
	err := r.db.QueryRow(query, id).Scan(
		&machine.ID,
		&machine.Name,
		&machine.Description,
		&machine.Host,
		&machine.Port,
		&machine.Username,
		&machine.AuthType,
		&machine.Status,
		&machine.CreatedAt,
		&machine.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get machine: %w", err)
	}

	return &machine, nil
}

// GetWithAuth retrieves a machine by ID including auth_value
func (r *SQLiteMachineRepository) GetWithAuth(id string) (*models.Machine, error) {
	query := `
	SELECT id, name, description, host, port, username, auth_type, auth_value, status, created_at, updated_at
	FROM machines
	WHERE id = ?
	`

	var machine models.Machine
	err := r.db.QueryRow(query, id).Scan(
		&machine.ID,
		&machine.Name,
		&machine.Description,
		&machine.Host,
		&machine.Port,
		&machine.Username,
		&machine.AuthType,
		&machine.AuthValue,
		&machine.Status,
		&machine.CreatedAt,
		&machine.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get machine with auth: %w", err)
	}

	return &machine, nil
}

// List retrieves all machines (without auth_value)
func (r *SQLiteMachineRepository) List() ([]*models.Machine, error) {
	query := `
	SELECT id, name, description, host, port, username, auth_type, status, created_at, updated_at
	FROM machines
	ORDER BY name ASC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list machines: %w", err)
	}
	defer rows.Close()

	var machines []*models.Machine
	for rows.Next() {
		var machine models.Machine
		err := rows.Scan(
			&machine.ID,
			&machine.Name,
			&machine.Description,
			&machine.Host,
			&machine.Port,
			&machine.Username,
			&machine.AuthType,
			&machine.Status,
			&machine.CreatedAt,
			&machine.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan machine: %w", err)
		}
		machines = append(machines, &machine)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating machines: %w", err)
	}

	return machines, nil
}

// Update updates an existing machine
func (r *SQLiteMachineRepository) Update(id, name, description, host string, port int, username, authType, encryptedAuthValue string) error {
	now := time.Now()

	query := `
	UPDATE machines
	SET name = ?, description = ?, host = ?, port = ?, username = ?, auth_type = ?, auth_value = ?, status = 'untested', updated_at = ?
	WHERE id = ?
	`

	result, err := r.db.Exec(query, name, description, host, port, username, authType, encryptedAuthValue, now, id)
	if err != nil {
		return fmt.Errorf("failed to update machine: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("machine not found: %s", id)
	}

	log.Printf("Updated machine: %s", id)
	return nil
}

// UpdateStatus updates the status of a machine
func (r *SQLiteMachineRepository) UpdateStatus(id, status string) error {
	now := time.Now()

	query := `
	UPDATE machines
	SET status = ?, updated_at = ?
	WHERE id = ?
	`

	result, err := r.db.Exec(query, status, now, id)
	if err != nil {
		return fmt.Errorf("failed to update machine status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("machine not found: %s", id)
	}

	log.Printf("Updated machine status: %s -> %s", id, status)
	return nil
}

// Delete deletes a machine
func (r *SQLiteMachineRepository) Delete(id string) error {
	result, err := r.db.Exec("DELETE FROM machines WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete machine: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("machine not found: %s", id)
	}

	log.Printf("Deleted machine: %s", id)
	return nil
}
