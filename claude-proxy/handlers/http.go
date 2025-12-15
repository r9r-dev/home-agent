package handlers

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/ronan/claude-proxy/services"
)

// HTTPHandler handles HTTP requests
type HTTPHandler struct {
	claudeService *services.ClaudeService
}

// NewHTTPHandler creates a new HTTPHandler
func NewHTTPHandler(claudeService *services.ClaudeService) *HTTPHandler {
	return &HTTPHandler{
		claudeService: claudeService,
	}
}

// TitleRequest represents a request to generate a title
type TitleRequest struct {
	UserMessage       string `json:"user_message"`
	AssistantResponse string `json:"assistant_response"`
}

// TitleResponse represents the response from title generation
type TitleResponse struct {
	Title string `json:"title"`
	Error string `json:"error,omitempty"`
}

// HealthCheck handles the health check endpoint
func (h *HTTPHandler) HealthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "healthy",
		"service": "claude-proxy",
	})
}

// GenerateTitle handles the title generation endpoint
func (h *HTTPHandler) GenerateTitle(c *fiber.Ctx) error {
	var req TitleRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("GenerateTitle: Failed to parse request: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(TitleResponse{
			Error: "Invalid request body",
		})
	}

	if req.UserMessage == "" || req.AssistantResponse == "" {
		return c.Status(fiber.StatusBadRequest).JSON(TitleResponse{
			Error: "user_message and assistant_response are required",
		})
	}

	title, err := h.claudeService.GenerateTitleSummary(req.UserMessage, req.AssistantResponse)
	if err != nil {
		log.Printf("GenerateTitle: Failed to generate title: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(TitleResponse{
			Error: err.Error(),
		})
	}

	return c.JSON(TitleResponse{
		Title: title,
	})
}
