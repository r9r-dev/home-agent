package models

import "time"

// Machine represents an SSH machine configuration
type Machine struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Host        string    `json:"host"`
	Port        int       `json:"port"`
	Username    string    `json:"username"`
	AuthType    string    `json:"auth_type"` // "password" or "key"
	AuthValue   string    `json:"-"`         // Encrypted, never returned in JSON
	Status      string    `json:"status"`    // "untested", "online", "offline"
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
