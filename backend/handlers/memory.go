package handlers

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/ronan/home-agent/models"
	"github.com/ronan/home-agent/repositories"
)

// MemoryHandler handles memory-related API endpoints
type MemoryHandler struct {
	memory repositories.MemoryRepository
}

// NewMemoryHandler creates a new MemoryHandler
func NewMemoryHandler(memory repositories.MemoryRepository) *MemoryHandler {
	return &MemoryHandler{memory: memory}
}

// RegisterRoutes registers memory API routes
func (h *MemoryHandler) RegisterRoutes(app *fiber.App) {
	app.Get("/api/memory", h.List)
	app.Post("/api/memory", h.Create)
	app.Get("/api/memory/export", h.Export)
	app.Post("/api/memory/import", h.Import)
	app.Get("/api/memory/:id", h.Get)
	app.Put("/api/memory/:id", h.Update)
	app.Delete("/api/memory/:id", h.Delete)
}

// CreateMemoryRequest represents the request body for creating a memory entry
type CreateMemoryRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// UpdateMemoryRequest represents the request body for updating a memory entry
type UpdateMemoryRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Enabled *bool  `json:"enabled,omitempty"`
}

// ImportMemoryRequest represents the request body for importing memory entries
type ImportMemoryRequest struct {
	Entries []struct {
		Title   string `json:"title"`
		Content string `json:"content"`
		Enabled bool   `json:"enabled"`
	} `json:"entries"`
}

// List returns all memory entries
func (h *MemoryHandler) List(c *fiber.Ctx) error {
	entries, err := h.memory.List()
	if err != nil {
		log.Printf("Failed to list memory entries: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list memory entries",
		})
	}

	// Return empty array if no entries
	if entries == nil {
		entries = []*models.MemoryEntry{}
	}

	return c.JSON(entries)
}

// Create creates a new memory entry
func (h *MemoryHandler) Create(c *fiber.Ctx) error {
	var req CreateMemoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate request
	if req.Title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Title is required",
		})
	}
	if req.Content == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Content is required",
		})
	}

	// Generate UUID for the entry
	id := uuid.New().String()

	entry, err := h.memory.Create(id, req.Title, req.Content)
	if err != nil {
		log.Printf("Failed to create memory entry: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create memory entry",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(entry)
}

// Get retrieves a single memory entry by ID
func (h *MemoryHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID is required",
		})
	}

	entry, err := h.memory.Get(id)
	if err != nil {
		log.Printf("Failed to get memory entry: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get memory entry",
		})
	}

	if entry == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Memory entry not found",
		})
	}

	return c.JSON(entry)
}

// Update updates an existing memory entry
func (h *MemoryHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID is required",
		})
	}

	var req UpdateMemoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Get existing entry to preserve values if not provided
	existing, err := h.memory.Get(id)
	if err != nil {
		log.Printf("Failed to get memory entry: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get memory entry",
		})
	}

	if existing == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Memory entry not found",
		})
	}

	// Use existing values if not provided
	title := req.Title
	if title == "" {
		title = existing.Title
	}
	content := req.Content
	if content == "" {
		content = existing.Content
	}
	enabled := existing.Enabled
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	err = h.memory.Update(id, title, content, enabled)
	if err != nil {
		log.Printf("Failed to update memory entry: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update memory entry",
		})
	}

	// Return updated entry
	entry, _ := h.memory.Get(id)
	return c.JSON(entry)
}

// Delete removes a memory entry
func (h *MemoryHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID is required",
		})
	}

	err := h.memory.Delete(id)
	if err != nil {
		log.Printf("Failed to delete memory entry: %v", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Memory entry not found",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// Export returns all memory entries in a format suitable for backup
func (h *MemoryHandler) Export(c *fiber.Ctx) error {
	entries, err := h.memory.List()
	if err != nil {
		log.Printf("Failed to export memory entries: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to export memory entries",
		})
	}

	// Return empty array if no entries
	if entries == nil {
		entries = []*models.MemoryEntry{}
	}

	// Create export format
	export := fiber.Map{
		"version": "1.0",
		"entries": entries,
	}

	return c.JSON(export)
}

// Import imports memory entries from a backup
func (h *MemoryHandler) Import(c *fiber.Ctx) error {
	var req ImportMemoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	imported := 0
	errors := []string{}

	for _, entry := range req.Entries {
		if entry.Title == "" || entry.Content == "" {
			errors = append(errors, "Skipped entry with empty title or content")
			continue
		}

		id := uuid.New().String()
		_, err := h.memory.Create(id, entry.Title, entry.Content)
		if err != nil {
			errors = append(errors, "Failed to import: "+entry.Title)
			continue
		}

		// Set enabled state if different from default
		if !entry.Enabled {
			h.memory.Update(id, entry.Title, entry.Content, false)
		}

		imported++
	}

	return c.JSON(fiber.Map{
		"imported": imported,
		"errors":   errors,
	})
}
