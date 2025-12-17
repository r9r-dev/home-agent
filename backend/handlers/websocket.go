package handlers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	chatHandler *ChatHandler
}

// NewWebSocketHandler creates a new WebSocketHandler instance
func NewWebSocketHandler(chatHandler *ChatHandler) *WebSocketHandler {
	return &WebSocketHandler{
		chatHandler: chatHandler,
	}
}

// Attachment represents a file attachment in a message
type Attachment struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
	Path     string `json:"path"`
	Type     string `json:"type"` // "image" or "file"
	MimeType string `json:"mime_type,omitempty"`
}

// ClientMessage represents a message from the WebSocket client
type ClientMessage struct {
	Type        string       `json:"type"`                  // "message", "ping", "history"
	Content     string       `json:"content,omitempty"`     // Message content
	SessionID   string       `json:"sessionId,omitempty"`   // Optional session ID
	Model       string       `json:"model,omitempty"`       // Claude model: haiku, sonnet, opus
	Attachments []Attachment `json:"attachments,omitempty"` // File attachments
	Thinking    bool         `json:"thinking,omitempty"`    // Enable extended thinking mode
}

// ServerMessage represents a message sent to the WebSocket client
type ServerMessage struct {
	Type      string `json:"type"`                // "chunk", "thinking", "done", "error", "pong", "history", "session_id"
	Content   string `json:"content,omitempty"`   // Message content
	SessionID string `json:"sessionId,omitempty"` // Session ID
	Error     string `json:"error,omitempty"`     // Error message
	Messages  []MessageResponse `json:"messages,omitempty"` // History messages
}

// UpgradeMiddleware checks if the request should be upgraded to WebSocket
func (wsh *WebSocketHandler) UpgradeMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check if it's a websocket upgrade request
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	}
}

// HandleWebSocket handles WebSocket connections
func (wsh *WebSocketHandler) HandleWebSocket(c *websocket.Conn) {
	clientAddr := c.RemoteAddr().String()

	// Set up connection parameters
	c.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.SetPongHandler(func(string) error {
		c.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// Start ping ticker
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Channel to signal when to close
	done := make(chan struct{})

	// Start goroutine to send pings
	go func() {
		for {
			select {
			case <-ticker.C:
				if err := c.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second)); err != nil {
					return
				}
			case <-done:
				return
			}
		}
	}()

	// Main message loop
	for {
		// Read message from client
		messageType, messageData, err := c.ReadMessage()
		if err != nil {
			break
		}

		// Only handle text messages
		if messageType != websocket.TextMessage {
			continue
		}

		// Reset read deadline
		c.SetReadDeadline(time.Now().Add(60 * time.Second))

		// Parse client message
		var clientMsg ClientMessage
		if err := json.Unmarshal(messageData, &clientMsg); err != nil {
			wsh.sendError(c, "Invalid JSON format")
			continue
		}

		// Handle different message types
		switch clientMsg.Type {
		case "message":
			wsh.handleChatMessage(c, clientMsg, clientAddr)

		case "ping":
			wsh.sendPong(c)

		case "history":
			wsh.handleHistory(c, clientMsg, clientAddr)

		default:
			wsh.sendError(c, "Unknown message type: "+clientMsg.Type)
		}
	}

	// Clean up
	close(done)
}

// handleChatMessage processes a chat message from the client
func (wsh *WebSocketHandler) handleChatMessage(c *websocket.Conn, clientMsg ClientMessage, clientAddr string) {
	// Convert attachments
	attachments := make([]MessageAttachment, len(clientMsg.Attachments))
	for i, a := range clientMsg.Attachments {
		attachments[i] = MessageAttachment{
			ID:       a.ID,
			Filename: a.Filename,
			Path:     a.Path,
			Type:     a.Type,
			MimeType: a.MimeType,
		}
	}

	// Create message request
	request := MessageRequest{
		Content:     clientMsg.Content,
		SessionID:   clientMsg.SessionID,
		Model:       clientMsg.Model,
		Attachments: attachments,
		Thinking:    clientMsg.Thinking,
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Handle the message
	responseChan, err := wsh.chatHandler.HandleMessage(ctx, request)
	if err != nil {
		wsh.sendError(c, err.Error())
		return
	}

	// Stream responses to client
	for response := range responseChan {
		var serverMsg ServerMessage

		switch response.Type {
		case "chunk":
			serverMsg = ServerMessage{
				Type:    "chunk",
				Content: response.Content,
			}

		case "thinking":
			serverMsg = ServerMessage{
				Type:    "thinking",
				Content: response.Content,
			}

		case "session_id":
			serverMsg = ServerMessage{
				Type:      "session_id",
				SessionID: response.SessionID,
			}

		case "done":
			serverMsg = ServerMessage{
				Type:      "done",
				SessionID: response.SessionID,
				Content:   response.Content,
			}

		case "error":
			serverMsg = ServerMessage{
				Type:  "error",
				Error: response.Error,
			}
		}

		// Send to client
		if err := wsh.sendMessage(c, serverMsg); err != nil {
			return
		}
	}
}

// handleHistory retrieves and sends conversation history
func (wsh *WebSocketHandler) handleHistory(c *websocket.Conn, clientMsg ClientMessage, clientAddr string) {
	if clientMsg.SessionID == "" {
		wsh.sendError(c, "Session ID is required for history")
		return
	}

	// Get history
	messages, err := wsh.chatHandler.GetHistory(clientMsg.SessionID)
	if err != nil {
		wsh.sendError(c, err.Error())
		return
	}

	// Send history response
	serverMsg := ServerMessage{
		Type:     "history",
		Messages: messages,
	}

	wsh.sendMessage(c, serverMsg)
}

// sendMessage sends a message to the WebSocket client
func (wsh *WebSocketHandler) sendMessage(c *websocket.Conn, msg ServerMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return c.WriteMessage(websocket.TextMessage, data)
}

// sendError sends an error message to the client
func (wsh *WebSocketHandler) sendError(c *websocket.Conn, errorMsg string) {
	msg := ServerMessage{
		Type:  "error",
		Error: errorMsg,
	}
	wsh.sendMessage(c, msg)
}

// sendPong sends a pong response to the client
func (wsh *WebSocketHandler) sendPong(c *websocket.Conn) {
	msg := ServerMessage{
		Type: "pong",
	}
	wsh.sendMessage(c, msg)
}

// RegisterRoutes registers WebSocket routes with the Fiber app
func (wsh *WebSocketHandler) RegisterRoutes(app *fiber.App) {
	// WebSocket upgrade middleware
	app.Use("/ws", wsh.UpgradeMiddleware())

	// WebSocket handler
	app.Get("/ws", websocket.New(wsh.HandleWebSocket))
}
