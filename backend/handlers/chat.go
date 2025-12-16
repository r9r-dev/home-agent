package handlers

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/ronan/home-agent/services"
)

// ChatHandler handles chat message processing and coordination
type ChatHandler struct {
	sessionManager  *services.SessionManager
	claudeExecutor  services.ClaudeExecutor
}

// NewChatHandler creates a new ChatHandler instance
func NewChatHandler(sessionManager *services.SessionManager, claudeExecutor services.ClaudeExecutor) *ChatHandler {
	log.Println("Initializing ChatHandler")
	return &ChatHandler{
		sessionManager:  sessionManager,
		claudeExecutor:  claudeExecutor,
	}
}

// MessageRequest represents an incoming message from the client
type MessageRequest struct {
	Content   string `json:"content"`
	SessionID string `json:"session_id,omitempty"`
	Model     string `json:"model,omitempty"` // Claude model: haiku, sonnet, opus
}

// MessageResponse represents a response chunk sent to the client
type MessageResponse struct {
	Type      string `json:"type"`       // "chunk", "done", "error", "session_id"
	Content   string `json:"content,omitempty"`
	SessionID string `json:"session_id,omitempty"`
	Error     string `json:"error,omitempty"`
}

// HandleMessage processes a user message and streams Claude's response
// It returns a channel that emits response chunks
func (ch *ChatHandler) HandleMessage(ctx context.Context, request MessageRequest) (<-chan MessageResponse, error) {
	// Default to haiku if no model specified
	model := request.Model
	if model == "" {
		model = "haiku"
	}

	log.Printf("HandleMessage: received message (length: %d), sessionID: %s, model: %s", len(request.Content), request.SessionID, model)

	// Validate input
	if strings.TrimSpace(request.Content) == "" {
		return nil, fmt.Errorf("message content cannot be empty")
	}

	// Get or create session with the specified model
	sessionID, isNew, err := ch.sessionManager.GetOrCreateSessionWithModel(request.SessionID, model)
	if err != nil {
		return nil, fmt.Errorf("failed to get or create session: %w", err)
	}

	if isNew {
		log.Printf("Created new session: %s with model: %s", sessionID, model)
	} else {
		log.Printf("Using existing session: %s", sessionID)
		// Update the model if the session already exists (user might have changed it)
		if err := ch.sessionManager.UpdateSessionModel(sessionID, model); err != nil {
			log.Printf("Warning: failed to update session model: %v", err)
		}
	}

	// Save user message to database
	if err := ch.sessionManager.SaveMessage(sessionID, "user", request.Content); err != nil {
		log.Printf("Warning: failed to save user message: %v", err)
		// Don't fail the request, just log the error
	}

	// Execute Claude command
	// For new sessions, don't pass a session ID to Claude - it will create one
	// For existing sessions, use the stored Claude session ID
	claudeSessionID := ""
	if !isNew {
		// Get the Claude session ID from database
		storedClaudeID, err := ch.sessionManager.GetClaudeSessionIDFromDB(sessionID)
		if err != nil {
			log.Printf("Warning: failed to get Claude session ID: %v", err)
		} else if storedClaudeID != "" {
			claudeSessionID = storedClaudeID
			log.Printf("Using stored Claude session ID: %s", claudeSessionID)
		}
	}
	claudeResponseChan, err := ch.claudeExecutor.ExecuteClaude(ctx, request.Content, claudeSessionID, model)
	if err != nil {
		return nil, fmt.Errorf("failed to execute claude: %w", err)
	}

	// Create response channel
	responseChan := make(chan MessageResponse, 100)

	// Start goroutine to process Claude's responses
	go ch.processClaudeResponse(sessionID, isNew, request.Content, claudeResponseChan, responseChan)

	return responseChan, nil
}

// processClaudeResponse processes the Claude response stream and sends formatted responses
func (ch *ChatHandler) processClaudeResponse(sessionID string, isNewSession bool, userMessage string, claudeResponseChan <-chan services.ClaudeResponse, responseChan chan<- MessageResponse) {
	defer close(responseChan)

	var fullAssistantResponse strings.Builder
	var finalSessionID string

	for claudeResp := range claudeResponseChan {
		switch claudeResp.Type {
		case "chunk":
			// Accumulate the full response
			fullAssistantResponse.WriteString(claudeResp.Content)

			// Send chunk to client
			responseChan <- MessageResponse{
				Type:    "chunk",
				Content: claudeResp.Content,
			}

		case "session_id":
			// Update session ID if provided by Claude
			if claudeResp.SessionID != "" {
				finalSessionID = claudeResp.SessionID
				log.Printf("Received session ID from Claude: %s", finalSessionID)

				// Save Claude's session ID to database for future resumes
				if err := ch.sessionManager.UpdateClaudeSessionID(sessionID, finalSessionID); err != nil {
					log.Printf("Warning: failed to save Claude session ID: %v", err)
				}

				// Send our internal session ID to client (not Claude's)
				responseChan <- MessageResponse{
					Type:      "session_id",
					SessionID: sessionID,
				}
			}

		case "done":
			log.Println("Claude response complete")

			// Use the session ID from the done event if available (for updating Claude session ID)
			if claudeResp.SessionID != "" && finalSessionID == "" {
				finalSessionID = claudeResp.SessionID
				// Save Claude's session ID to database for future resumes
				if err := ch.sessionManager.UpdateClaudeSessionID(sessionID, finalSessionID); err != nil {
					log.Printf("Warning: failed to save Claude session ID: %v", err)
				}
			}

			// Save assistant's full response to database using our internal session ID
			assistantMessage := fullAssistantResponse.String()
			if assistantMessage != "" {
				if err := ch.sessionManager.SaveMessage(sessionID, "assistant", assistantMessage); err != nil {
					log.Printf("Warning: failed to save assistant message: %v", err)
				}
			}

			// Send done signal to client with our internal session ID
			responseChan <- MessageResponse{
				Type:      "done",
				SessionID: sessionID,
				Content:   assistantMessage,
			}

			// Generate a summary title for new sessions (async, don't block)
			if isNewSession && assistantMessage != "" {
				go func() {
					title, err := ch.claudeExecutor.GenerateTitleSummary(userMessage, assistantMessage)
					if err != nil {
						log.Printf("Warning: failed to generate title summary: %v", err)
						return
					}
					if title != "" {
						if err := ch.sessionManager.UpdateSessionTitle(sessionID, title); err != nil {
							log.Printf("Warning: failed to update session title: %v", err)
						} else {
							log.Printf("Updated session title to: %s", title)
						}
					}
				}()
			}

		case "error":
			log.Printf("Error from Claude: %v", claudeResp.Error)

			// Send error to client
			responseChan <- MessageResponse{
				Type:  "error",
				Error: claudeResp.Error.Error(),
			}
			return
		}
	}
}

// GetHistory retrieves the conversation history for a session
func (ch *ChatHandler) GetHistory(sessionID string) ([]MessageResponse, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session ID is required")
	}

	// Check if session exists
	if !ch.sessionManager.SessionExists(sessionID) {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	// Get messages from database
	messages, err := ch.sessionManager.GetMessages(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	// Convert to response format
	responses := make([]MessageResponse, len(messages))
	for i, msg := range messages {
		responses[i] = MessageResponse{
			Type:      "message",
			Content:   msg.Content,
			SessionID: msg.SessionID,
		}
	}

	return responses, nil
}

// ValidateSession checks if a session exists and is valid
func (ch *ChatHandler) ValidateSession(sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("session ID is required")
	}

	if !ch.sessionManager.SessionExists(sessionID) {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	return nil
}
