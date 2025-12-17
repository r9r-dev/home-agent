package handlers

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ronan/home-agent/models"
	"github.com/ronan/home-agent/services"
)

// ChatHandler handles chat message processing and coordination
type ChatHandler struct {
	sessionManager *services.SessionManager
	claudeExecutor services.ClaudeExecutor
	uploadDir      string // Local path for file storage
	workspacePath  string // Path prefix for Claude CLI (if different from uploadDir)
	db             *models.DB
}

// NewChatHandler creates a new ChatHandler instance
func NewChatHandler(sessionManager *services.SessionManager, claudeExecutor services.ClaudeExecutor, uploadDir string, workspacePath string, db *models.DB) *ChatHandler {
	log.Println("Initializing ChatHandler")
	return &ChatHandler{
		sessionManager: sessionManager,
		claudeExecutor: claudeExecutor,
		uploadDir:      uploadDir,
		workspacePath:  workspacePath,
		db:             db,
	}
}

// MessageAttachment represents a file attached to a message
type MessageAttachment struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
	Path     string `json:"path"`
	Type     string `json:"type"` // "image" or "file"
	MimeType string `json:"mime_type,omitempty"`
}

// MessageRequest represents an incoming message from the client
type MessageRequest struct {
	Content     string              `json:"content"`
	SessionID   string              `json:"session_id,omitempty"`
	Model       string              `json:"model,omitempty"` // Claude model: haiku, sonnet, opus
	Attachments []MessageAttachment `json:"attachments,omitempty"`
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

	log.Printf("HandleMessage: received message (length: %d), sessionID: %s, model: %s, attachments: %d", len(request.Content), request.SessionID, model, len(request.Attachments))

	// Validate input - allow empty content if attachments are present
	if strings.TrimSpace(request.Content) == "" && len(request.Attachments) == 0 {
		return nil, fmt.Errorf("message content cannot be empty")
	}

	// Build prompt with attachments
	prompt := ch.buildPromptWithAttachments(request.Content, request.Attachments)

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

	// Save user message to database (save original content, not the augmented prompt)
	userContent := request.Content
	if len(request.Attachments) > 0 {
		// Include attachment info in saved message for display
		userContent = ch.buildDisplayContentWithAttachments(request.Content, request.Attachments)
	}
	if err := ch.sessionManager.SaveMessage(sessionID, "user", userContent); err != nil {
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

	// Get custom instructions from settings
	customInstructions := ""
	if ch.db != nil {
		if instructions, err := ch.db.GetSetting("custom_instructions"); err == nil {
			customInstructions = instructions
		} else {
			log.Printf("Warning: failed to get custom instructions: %v", err)
		}
	}

	// Get enabled memory entries and format them
	memoryContext := ""
	if ch.db != nil {
		if entries, err := ch.db.GetEnabledMemoryEntries(); err == nil && len(entries) > 0 {
			// Convert to services.MemoryEntry format
			memoryEntries := make([]services.MemoryEntry, len(entries))
			for i, e := range entries {
				memoryEntries[i] = services.MemoryEntry{
					Title:   e.Title,
					Content: e.Content,
				}
			}
			memoryContext = services.FormatMemoryEntries(memoryEntries)
			log.Printf("Injecting %d memory entries into prompt", len(entries))
		}
	}

	// Combine memory and custom instructions
	fullInstructions := customInstructions
	if memoryContext != "" {
		if fullInstructions != "" {
			fullInstructions = memoryContext + "\n\n" + fullInstructions
		} else {
			fullInstructions = memoryContext
		}
	}

	claudeResponseChan, err := ch.claudeExecutor.ExecuteClaude(ctx, prompt, claudeSessionID, model, fullInstructions)
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

// buildPromptWithAttachments builds a prompt that includes attachment content for Claude
func (ch *ChatHandler) buildPromptWithAttachments(content string, attachments []MessageAttachment) string {
	if len(attachments) == 0 {
		return content
	}

	var sb strings.Builder

	// Process each attachment
	for _, att := range attachments {
		physicalPath := ch.getPhysicalPath(att.Path)
		if physicalPath == "" {
			log.Printf("Warning: could not resolve physical path for %s", att.Path)
			continue
		}

		if att.Type == "image" {
			// For images, provide the path for Claude to read
			claudePath := ch.getClaudePath(physicalPath)
			sb.WriteString(fmt.Sprintf("[Image: %s]\nPlease read and analyze this image file: %s\n\n", att.Filename, claudePath))
		} else {
			// For text files, read and include the content directly
			fileContent, err := ch.readFileContent(physicalPath)
			if err != nil {
				log.Printf("Warning: could not read file %s: %v", physicalPath, err)
				sb.WriteString(fmt.Sprintf("[File: %s - Error reading content]\n\n", att.Filename))
				continue
			}
			sb.WriteString(fmt.Sprintf("[File: %s]\n```\n%s\n```\n\n", att.Filename, fileContent))
		}
	}

	// Add the user's message
	if content != "" {
		sb.WriteString(content)
	} else {
		sb.WriteString("Please analyze the attached file(s).")
	}

	return sb.String()
}

// getClaudePath returns the path that Claude CLI should use to access the file
// If workspacePath is set, maps the local path to the workspace path
// Otherwise returns the absolute local path
func (ch *ChatHandler) getClaudePath(localPath string) string {
	if ch.workspacePath != "" {
		// Extract the relative path from uploadDir
		// localPath: /data/uploads/temp/xxx-file.png
		// uploadDir: /data/uploads
		// relPath: temp/xxx-file.png
		relPath, err := filepath.Rel(ch.uploadDir, localPath)
		if err != nil {
			log.Printf("Warning: could not get relative path for %s: %v", localPath, err)
			return localPath
		}
		// Build the Claude path: workspacePath/uploads/relPath
		// workspacePath: /home/user/workspace
		// result: /home/user/workspace/uploads/temp/xxx-file.png
		claudePath := filepath.Join(ch.workspacePath, "uploads", relPath)
		log.Printf("Mapped local path %s to Claude path %s", localPath, claudePath)
		return claudePath
	}

	// No workspace mapping, use absolute local path
	absPath, err := filepath.Abs(localPath)
	if err != nil {
		log.Printf("Warning: could not get absolute path for %s: %v", localPath, err)
		return localPath
	}
	return absPath
}

// getPhysicalPath converts an API path to a physical file path
// API path format: /api/uploads/{session_id}/{filename}
func (ch *ChatHandler) getPhysicalPath(apiPath string) string {
	// Remove /api/uploads/ prefix
	prefix := "/api/uploads/"
	if !strings.HasPrefix(apiPath, prefix) {
		return ""
	}

	relativePath := strings.TrimPrefix(apiPath, prefix)
	return filepath.Join(ch.uploadDir, relativePath)
}

// readFileContent reads the content of a text file
func (ch *ChatHandler) readFileContent(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Limit reading to 100KB to avoid huge prompts
	const maxSize = 100 * 1024
	limitedReader := io.LimitReader(file, maxSize)

	content, err := io.ReadAll(limitedReader)
	if err != nil {
		return "", err
	}

	result := string(content)

	// Check if file was truncated
	stat, err := file.Stat()
	if err == nil && stat.Size() > maxSize {
		result += "\n... [file truncated, showing first 100KB]"
	}

	return result, nil
}

// buildDisplayContentWithAttachments builds content for display that shows attachment info
func (ch *ChatHandler) buildDisplayContentWithAttachments(content string, attachments []MessageAttachment) string {
	if len(attachments) == 0 {
		return content
	}

	var sb strings.Builder

	// Add the user's message first
	if content != "" {
		sb.WriteString(content)
		sb.WriteString("\n\n")
	}

	// Add attachment markers for frontend to parse
	sb.WriteString("<!-- attachments:")
	for i, att := range attachments {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(fmt.Sprintf("%s|%s|%s|%s", att.ID, att.Filename, att.Path, att.Type))
	}
	sb.WriteString(" -->")

	return sb.String()
}
