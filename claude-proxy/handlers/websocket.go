package handlers

import (
	"context"
	"log"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/ronan/claude-proxy/services"
)

// WebSocketHandler handles WebSocket connections for Claude execution
type WebSocketHandler struct {
	claudeService *services.ClaudeService
	apiKey        string
}

// NewWebSocketHandler creates a new WebSocketHandler
func NewWebSocketHandler(claudeService *services.ClaudeService, apiKey string) *WebSocketHandler {
	return &WebSocketHandler{
		claudeService: claudeService,
		apiKey:        apiKey,
	}
}

// ProxyRequest represents a request from the client
type ProxyRequest struct {
	Type               string `json:"type"`
	Prompt             string `json:"prompt"`
	SessionID          string `json:"session_id,omitempty"`
	IsNewSession       bool   `json:"is_new_session,omitempty"`
	Model              string `json:"model,omitempty"`
	CustomInstructions string `json:"custom_instructions,omitempty"`
	Thinking           bool   `json:"thinking,omitempty"`
}

// ProxyResponse represents a response to the client
type ProxyResponse struct {
	Type      string `json:"type"`
	Content   string `json:"content,omitempty"`
	SessionID string `json:"session_id,omitempty"`
	Error     string `json:"error,omitempty"`
}

// RegisterRoutes registers WebSocket routes
func (wsh *WebSocketHandler) RegisterRoutes(app *fiber.App) {
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			// Validate API key for WebSocket upgrade
			if wsh.apiKey != "" {
				providedKey := c.Get("X-API-Key")
				if providedKey != wsh.apiKey {
					log.Printf("WebSocket: Invalid API key from %s", c.IP())
					return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
						"error": "Invalid API key",
					})
				}
			}
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws", websocket.New(wsh.HandleConnection))
}

// HandleConnection handles a WebSocket connection
func (wsh *WebSocketHandler) HandleConnection(c *websocket.Conn) {
	log.Printf("WebSocket: New connection from %s", c.RemoteAddr())
	defer func() {
		log.Printf("WebSocket: Connection closed from %s", c.RemoteAddr())
		c.Close()
	}()

	// Handle messages
	for {
		var request ProxyRequest
		if err := c.ReadJSON(&request); err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				log.Println("WebSocket: Client disconnected normally")
				return
			}
			log.Printf("WebSocket: Failed to read message: %v", err)
			return
		}

		log.Printf("WebSocket: Received request type=%s, prompt_len=%d, session_id=%s, is_new_session=%v, model=%s, thinking=%v",
			request.Type, len(request.Prompt), request.SessionID, request.IsNewSession, request.Model, request.Thinking)

		switch request.Type {
		case "execute":
			wsh.handleExecute(c, request)
		default:
			log.Printf("WebSocket: Unknown request type: %s", request.Type)
			wsh.sendError(c, "Unknown request type")
		}
	}
}

// handleExecute handles Claude execution requests
func (wsh *WebSocketHandler) handleExecute(c *websocket.Conn, request ProxyRequest) {
	if request.Prompt == "" {
		wsh.sendError(c, "Prompt is required")
		return
	}

	ctx := context.Background()

	// Execute Claude with session management
	responseChan, err := wsh.claudeService.ExecuteClaude(ctx, request.Prompt, request.SessionID, request.IsNewSession, request.Model, request.CustomInstructions, request.Thinking)
	if err != nil {
		log.Printf("WebSocket: Failed to execute Claude: %v", err)
		wsh.sendError(c, err.Error())
		return
	}

	// Use a mutex to protect WebSocket writes
	var writeMu sync.Mutex

	// Stream responses
	for response := range responseChan {
		var proxyResp ProxyResponse

		switch response.Type {
		case "chunk":
			proxyResp = ProxyResponse{
				Type:    "chunk",
				Content: response.Content,
			}

		case "thinking":
			proxyResp = ProxyResponse{
				Type:    "thinking",
				Content: response.Content,
			}

		case "session_id":
			proxyResp = ProxyResponse{
				Type:      "session_id",
				SessionID: response.SessionID,
			}

		case "done":
			proxyResp = ProxyResponse{
				Type:      "done",
				Content:   response.Content,
				SessionID: response.SessionID,
			}

		case "error":
			errMsg := "unknown error"
			if response.Error != nil {
				errMsg = response.Error.Error()
			}
			proxyResp = ProxyResponse{
				Type:  "error",
				Error: errMsg,
			}
		}

		writeMu.Lock()
		if err := c.WriteJSON(proxyResp); err != nil {
			writeMu.Unlock()
			log.Printf("WebSocket: Failed to send response: %v", err)
			return
		}
		writeMu.Unlock()

		// Exit after done or error
		if response.Type == "done" || response.Type == "error" {
			return
		}
	}
}

// sendError sends an error response
func (wsh *WebSocketHandler) sendError(c *websocket.Conn, message string) {
	if err := c.WriteJSON(ProxyResponse{
		Type:  "error",
		Error: message,
	}); err != nil {
		log.Printf("WebSocket: Failed to send error: %v", err)
	}
}
