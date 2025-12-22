package handlers

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/ronan/home-agent/models"
	"github.com/ronan/home-agent/repositories"
	"github.com/ronan/home-agent/services"
)

// MachinesHandler handles machine-related API endpoints
type MachinesHandler struct {
	machines repositories.MachineRepository
	crypto   *services.CryptoService
}

// NewMachinesHandler creates a new MachinesHandler
func NewMachinesHandler(machines repositories.MachineRepository, crypto *services.CryptoService) *MachinesHandler {
	return &MachinesHandler{machines: machines, crypto: crypto}
}

// RegisterRoutes registers machine API routes
func (h *MachinesHandler) RegisterRoutes(app *fiber.App) {
	app.Get("/api/machines", h.List)
	app.Post("/api/machines", h.Create)
	app.Get("/api/machines/:id", h.Get)
	app.Put("/api/machines/:id", h.Update)
	app.Delete("/api/machines/:id", h.Delete)
	app.Post("/api/machines/:id/test", h.TestConnection)
}

// CreateMachineRequest represents the request body for creating a machine
type CreateMachineRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Username    string `json:"username"`
	AuthType    string `json:"auth_type"`
	AuthValue   string `json:"auth_value"`
}

// UpdateMachineRequest represents the request body for updating a machine
type UpdateMachineRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Username    string `json:"username"`
	AuthType    string `json:"auth_type"`
	AuthValue   string `json:"auth_value"`
}

// List returns all machines
func (h *MachinesHandler) List(c *fiber.Ctx) error {
	machines, err := h.machines.List()
	if err != nil {
		log.Printf("Failed to list machines: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list machines",
		})
	}

	// Return empty array if no machines
	if machines == nil {
		machines = []*models.Machine{}
	}

	return c.JSON(machines)
}

// Create creates a new machine
func (h *MachinesHandler) Create(c *fiber.Ctx) error {
	var req CreateMachineRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate request
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name is required",
		})
	}
	if req.Host == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Host is required",
		})
	}
	if req.Username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Username is required",
		})
	}
	if req.AuthType != "password" && req.AuthType != "key" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Auth type must be 'password' or 'key'",
		})
	}
	if req.AuthValue == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Auth value is required",
		})
	}

	// Default port
	port := req.Port
	if port == 0 {
		port = 22
	}

	// Encrypt auth value
	encryptedAuthValue, err := h.crypto.Encrypt(req.AuthValue)
	if err != nil {
		log.Printf("Failed to encrypt auth value: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to encrypt credentials",
		})
	}

	// Generate UUID
	id := uuid.New().String()

	machine, err := h.machines.Create(id, req.Name, req.Description, req.Host, port, req.Username, req.AuthType, encryptedAuthValue)
	if err != nil {
		log.Printf("Failed to create machine: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create machine",
		})
	}

	// Clear auth value before returning
	machine.AuthValue = ""

	return c.Status(fiber.StatusCreated).JSON(machine)
}

// Get retrieves a single machine by ID
func (h *MachinesHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID is required",
		})
	}

	machine, err := h.machines.Get(id)
	if err != nil {
		log.Printf("Failed to get machine: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get machine",
		})
	}

	if machine == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Machine not found",
		})
	}

	return c.JSON(machine)
}

// Update updates an existing machine
func (h *MachinesHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID is required",
		})
	}

	var req UpdateMachineRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate request
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name is required",
		})
	}
	if req.Host == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Host is required",
		})
	}
	if req.Username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Username is required",
		})
	}
	if req.AuthType != "password" && req.AuthType != "key" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Auth type must be 'password' or 'key'",
		})
	}
	if req.AuthValue == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Auth value is required",
		})
	}

	// Default port
	port := req.Port
	if port == 0 {
		port = 22
	}

	// Encrypt auth value
	encryptedAuthValue, err := h.crypto.Encrypt(req.AuthValue)
	if err != nil {
		log.Printf("Failed to encrypt auth value: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to encrypt credentials",
		})
	}

	err = h.machines.Update(id, req.Name, req.Description, req.Host, port, req.Username, req.AuthType, encryptedAuthValue)
	if err != nil {
		log.Printf("Failed to update machine: %v", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Machine not found",
		})
	}

	// Return updated machine
	machine, _ := h.machines.Get(id)
	return c.JSON(machine)
}

// Delete removes a machine
func (h *MachinesHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID is required",
		})
	}

	err := h.machines.Delete(id)
	if err != nil {
		log.Printf("Failed to delete machine: %v", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Machine not found",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// TestConnection tests SSH connectivity to a machine
func (h *MachinesHandler) TestConnection(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID is required",
		})
	}

	// Get machine with auth value
	machine, err := h.machines.GetWithAuth(id)
	if err != nil {
		log.Printf("Failed to get machine: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get machine",
		})
	}

	if machine == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Machine not found",
		})
	}

	// Decrypt auth value
	decryptedAuthValue, err := h.crypto.Decrypt(machine.AuthValue)
	if err != nil {
		log.Printf("Failed to decrypt auth value: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to decrypt credentials",
		})
	}

	// Test connection
	result := services.TestSSHConnection(machine.Host, machine.Port, machine.Username, machine.AuthType, decryptedAuthValue)

	// Update machine status based on result
	status := "offline"
	if result.Success {
		status = "online"
	}
	if err := h.machines.UpdateStatus(id, status); err != nil {
		log.Printf("Failed to update machine status: %v", err)
	}

	return c.JSON(result)
}
