package handlers

import (
	"context"
	"encoding/json"
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
	logService     *services.LogService
	cryptoService  *services.CryptoService
}

// NewChatHandler creates a new ChatHandler instance
func NewChatHandler(sessionManager *services.SessionManager, claudeExecutor services.ClaudeExecutor, uploadDir string, workspacePath string, db *models.DB, logService *services.LogService, cryptoService *services.CryptoService) *ChatHandler {
	return &ChatHandler{
		sessionManager: sessionManager,
		claudeExecutor: claudeExecutor,
		uploadDir:      uploadDir,
		workspacePath:  workspacePath,
		db:             db,
		logService:     logService,
		cryptoService:  cryptoService,
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
	Thinking    bool                `json:"thinking,omitempty"` // Enable extended thinking mode
	MachineID   string              `json:"machine_id,omitempty"` // Target SSH machine ID
}

// ToolInfo represents tool information for WebSocket responses
type ToolInfo struct {
	ToolUseID  string                 `json:"tool_use_id"`
	ToolName   string                 `json:"tool_name"`
	Input      map[string]interface{} `json:"input,omitempty"`
}

// MessageResponse represents a response chunk sent to the client
type MessageResponse struct {
	Type      string `json:"type"`       // "chunk", "thinking", "thinking_end", "done", "error", "session_id", "session_title", "tool_start", "tool_progress", "tool_result", "tool_error", "tool_input_delta"
	Content   string `json:"content,omitempty"`
	SessionID string `json:"session_id,omitempty"`
	Title     string `json:"title,omitempty"` // Session title for session_title type
	Error     string `json:"error,omitempty"`
	// Tool-specific fields
	Tool               *ToolInfo `json:"tool,omitempty"`
	ElapsedTimeSeconds float64   `json:"elapsed_time_seconds,omitempty"`
	ToolOutput         string    `json:"tool_output,omitempty"`
	IsError            bool      `json:"is_error,omitempty"`
	InputDelta         string    `json:"input_delta,omitempty"` // JSON delta for streaming tool input
}

// HandleMessage processes a user message and streams Claude's response
// It returns a channel that emits response chunks
func (ch *ChatHandler) HandleMessage(ctx context.Context, request MessageRequest) (<-chan MessageResponse, error) {
	// Default to haiku if no model specified
	model := request.Model
	if model == "" {
		model = "haiku"
	}

	// Validate input - allow empty content if attachments are present
	if strings.TrimSpace(request.Content) == "" && len(request.Attachments) == 0 {
		return nil, fmt.Errorf("message content cannot be empty")
	}

	// Build prompt with attachments
	prompt := ch.buildPromptWithAttachments(request.Content, request.Attachments)

	// Prepare user content for saving (original content, not the augmented prompt)
	userContent := request.Content
	if len(request.Attachments) > 0 {
		// Include attachment info in saved message for display
		userContent = ch.buildDisplayContentWithAttachments(request.Content, request.Attachments)
	}

	// Determine if this is a new conversation or a resume
	isNewConversation := request.SessionID == ""
	sessionID := request.SessionID

	// For existing sessions, verify it exists and save user message now
	if !isNewConversation {
		if !ch.sessionManager.SessionExists(sessionID) {
			return nil, fmt.Errorf("session not found: %s", sessionID)
		}
		// Update model if changed
		ch.sessionManager.UpdateSessionModel(sessionID, model)
		// Save user message before calling SDK (it will be updated with new session_id)
		ch.sessionManager.SaveMessage(sessionID, "user", userContent)
	}

	// Get custom instructions from settings
	customInstructions := ""
	if ch.db != nil {
		if instructions, err := ch.db.GetSetting("custom_instructions"); err == nil {
			customInstructions = instructions
		}
	}

	// Get enabled memory entries and format them
	memoryContext := ""
	if ch.db != nil {
		if entries, err := ch.db.GetEnabledMemoryEntries(); err == nil && len(entries) > 0 {
			memoryEntries := make([]services.MemoryEntry, len(entries))
			for i, e := range entries {
				memoryEntries[i] = services.MemoryEntry{
					Title:   e.Title,
					Content: e.Content,
				}
			}
			memoryContext = services.FormatMemoryEntries(memoryEntries)
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

	// Add SSH machine context if machineId is provided
	if request.MachineID != "" && ch.db != nil && ch.cryptoService != nil {
		machine, err := ch.db.GetMachineWithAuth(request.MachineID)
		if err != nil {
			return nil, fmt.Errorf("machine not found: %s", request.MachineID)
		}

		// Decrypt auth value
		authValue, err := ch.cryptoService.Decrypt(machine.AuthValue)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt machine credentials: %w", err)
		}

		// Build SSH context
		sshContext := fmt.Sprintf(`<ssh_machine>
name: %s
description: %s
host: %s
port: %d
username: %s
auth_type: %s
</ssh_machine>

Tu as acces a cette machine SSH distante. Pour executer des commandes sur cette machine, utilise SSH:
- Si auth_type est "password": ssh %s@%s -p %d avec le mot de passe: %s
- Si auth_type est "key": utilise la cle SSH fournie avec ssh -i <key_file> %s@%s -p %d

La cle ou le mot de passe est: %s

IMPORTANT: Execute les commandes sur cette machine distante, pas en local.`,
			machine.Name,
			machine.Description,
			machine.Host,
			machine.Port,
			machine.Username,
			machine.AuthType,
			machine.Username, machine.Host, machine.Port, authValue,
			machine.Username, machine.Host, machine.Port,
			authValue,
		)

		if fullInstructions != "" {
			fullInstructions = sshContext + "\n\n" + fullInstructions
		} else {
			fullInstructions = sshContext
		}
	}

	// Execute Claude
	// For new conversations: sessionID is empty, SDK will generate one
	// For resume: sessionID is provided, SDK will resume and return new ID
	claudeResponseChan, err := ch.claudeExecutor.ExecuteClaude(ctx, prompt, sessionID, isNewConversation, model, fullInstructions, request.Thinking)
	if err != nil {
		return nil, fmt.Errorf("failed to execute claude: %w", err)
	}

	// Create response channel
	responseChan := make(chan MessageResponse, 100)

	// Start goroutine to process Claude's responses
	go ch.processClaudeResponse(sessionID, isNewConversation, model, userContent, request.Content, claudeResponseChan, responseChan)

	return responseChan, nil
}

// processClaudeResponse processes the Claude response stream and sends formatted responses
// oldSessionID: the session ID provided by frontend (empty for new conversations)
// isNewConversation: true if this is a new conversation
// model: the model to use
// userContent: the user message content to save (with attachment markers)
// userMessage: the original user message (for title generation)
func (ch *ChatHandler) processClaudeResponse(oldSessionID string, isNewConversation bool, model string, userContent string, userMessage string, claudeResponseChan <-chan services.ClaudeResponse, responseChan chan<- MessageResponse) {
	// Note: We don't defer close here because we need to send session_title after done
	// The channel will be closed at the end of this function

	var fullAssistantResponse strings.Builder
	var currentThinkingContent strings.Builder // Current thinking block being accumulated
	var currentSessionID string = oldSessionID
	var pendingTitleGeneration bool = false
	var titleUserMessage, titleAssistantMessage string
	var wasThinking bool = false // Track if we were receiving thinking content

	// Helper function to finalize current thinking block
	finalizeThinkingBlock := func() {
		if currentThinkingContent.Len() > 0 {
			thinkingContent := currentThinkingContent.String()
			// Send thinking_end to client
			responseChan <- MessageResponse{
				Type: "thinking_end",
			}
			// Save thinking content to database
			if currentSessionID != "" {
				if err := ch.sessionManager.SaveMessage(currentSessionID, "thinking", thinkingContent); err != nil {
					ch.logService.Error(fmt.Sprintf("Failed to save thinking message: %v", err))
				}
			}
			// Reset the thinking content builder
			currentThinkingContent.Reset()
		}
		wasThinking = false
	}

	for claudeResp := range claudeResponseChan {
		switch claudeResp.Type {
		case "chunk":
			// If we were thinking, finalize that block first
			if wasThinking {
				finalizeThinkingBlock()
			}

			// Accumulate the full response
			fullAssistantResponse.WriteString(claudeResp.Content)

			// Send chunk to client
			responseChan <- MessageResponse{
				Type:    "chunk",
				Content: claudeResp.Content,
			}

		case "thinking":
			// Accumulate thinking content
			currentThinkingContent.WriteString(claudeResp.Content)
			wasThinking = true

			// Send thinking content to client
			responseChan <- MessageResponse{
				Type:    "thinking",
				Content: claudeResp.Content,
			}

		case "session_id":
			// SDK returned a session_id
			sdkSessionID := claudeResp.SessionID

			if isNewConversation {
				// New conversation: create session with SDK's session_id
				_, err := ch.sessionManager.CreateSessionWithID(sdkSessionID, model)
				if err != nil {
					ch.logService.Error(fmt.Sprintf("Failed to create session: %v", err))
					responseChan <- MessageResponse{
						Type:  "error",
						Error: fmt.Sprintf("failed to create session: %v", err),
					}
					return
				}
				// Save user message now that we have a session
				if err := ch.sessionManager.SaveMessage(sdkSessionID, "user", userContent); err != nil {
					ch.logService.Error(fmt.Sprintf("Failed to save user message: %v", err))
				}
				currentSessionID = sdkSessionID
			} else {
				// Resume: update session_id if it changed
				if sdkSessionID != oldSessionID {
					err := ch.sessionManager.UpdateSessionID(oldSessionID, sdkSessionID)
					if err != nil {
						ch.logService.Warning(fmt.Sprintf("Failed to update session ID: %v", err))
					}
					currentSessionID = sdkSessionID
				}
			}

			// Send session_id to client (always use the current/new ID)
			responseChan <- MessageResponse{
				Type:      "session_id",
				SessionID: currentSessionID,
			}

		case "done":
			// Finalize any pending thinking block
			if wasThinking {
				finalizeThinkingBlock()
			}

			// Save assistant's full response to database
			assistantMessage := fullAssistantResponse.String()
			if assistantMessage != "" {
				if err := ch.sessionManager.SaveMessage(currentSessionID, "assistant", assistantMessage); err != nil {
					ch.logService.Error(fmt.Sprintf("Failed to save assistant message: %v", err))
				}
			}

			// Send done signal to client
			responseChan <- MessageResponse{
				Type:      "done",
				SessionID: currentSessionID,
				Content:   assistantMessage,
			}

			// Mark for title generation (will be done after the loop)
			if isNewConversation && assistantMessage != "" {
				pendingTitleGeneration = true
				titleUserMessage = userMessage
				titleAssistantMessage = assistantMessage
			}

		case "error":
			ch.logService.Error(fmt.Sprintf("Claude error: %v", claudeResp.Error))
			responseChan <- MessageResponse{
				Type:  "error",
				Error: claudeResp.Error.Error(),
			}
			return

		case "tool_start":
			// If we were thinking, finalize that block first
			if wasThinking {
				finalizeThinkingBlock()
			}

			log.Printf("[Chat] Received tool_start: %s (%s)", claudeResp.Tool.ToolName, claudeResp.Tool.ToolUseID)
			// Create tool call in database
			if claudeResp.Tool != nil && currentSessionID != "" {
				inputJSON, _ := json.Marshal(claudeResp.Tool.Input)
				_, err := ch.db.CreateToolCall(currentSessionID, claudeResp.Tool.ToolUseID, claudeResp.Tool.ToolName, string(inputJSON))
				if err != nil {
					log.Printf("Warning: failed to create tool call: %v", err)
				}
			}

			// Forward to client
			var toolInfo *ToolInfo
			if claudeResp.Tool != nil {
				toolInfo = &ToolInfo{
					ToolUseID: claudeResp.Tool.ToolUseID,
					ToolName:  claudeResp.Tool.ToolName,
					Input:     claudeResp.Tool.Input,
				}
			}
			responseChan <- MessageResponse{
				Type: "tool_start",
				Tool: toolInfo,
			}

		case "tool_progress":
			log.Printf("[Chat] Received tool_progress for tool %s (%.1fs)", claudeResp.Tool.ToolUseID, claudeResp.ElapsedTimeSeconds)
			// Forward progress to client (no DB update needed)
			var toolInfo *ToolInfo
			if claudeResp.Tool != nil {
				toolInfo = &ToolInfo{
					ToolUseID: claudeResp.Tool.ToolUseID,
					ToolName:  claudeResp.Tool.ToolName,
				}
			}
			responseChan <- MessageResponse{
				Type:               "tool_progress",
				Tool:               toolInfo,
				ElapsedTimeSeconds: claudeResp.ElapsedTimeSeconds,
			}

		case "tool_input_delta":
			log.Printf("[Chat] Received tool_input_delta for tool %s", claudeResp.Tool.ToolUseID)
			// Forward input delta to client for real-time display
			var toolInfo *ToolInfo
			if claudeResp.Tool != nil {
				toolInfo = &ToolInfo{
					ToolUseID: claudeResp.Tool.ToolUseID,
					ToolName:  claudeResp.Tool.ToolName,
				}
			}
			responseChan <- MessageResponse{
				Type:       "tool_input_delta",
				Tool:       toolInfo,
				InputDelta: claudeResp.InputDelta,
			}

		case "tool_result", "tool_error":
			log.Printf("[Chat] Received %s for tool %s", claudeResp.Type, claudeResp.Tool.ToolUseID)
			// Update tool call in database with input and result
			var inputMap map[string]interface{}
			if claudeResp.Tool != nil {
				status := "success"
				if claudeResp.IsError || claudeResp.Type == "tool_error" {
					status = "error"
				}
				// Convert input map to JSON string
				inputJSON, _ := json.Marshal(claudeResp.Tool.Input)
				inputMap = claudeResp.Tool.Input
				log.Printf("[Chat] Tool %s input: %s", claudeResp.Tool.ToolUseID, string(inputJSON))
				err := ch.db.UpdateToolCallOutput(claudeResp.Tool.ToolUseID, string(inputJSON), claudeResp.ToolOutput, status)
				if err != nil {
					log.Printf("Warning: failed to update tool call: %v", err)
				}
			}

			// Forward to client (include input for frontend to display)
			var toolInfo *ToolInfo
			if claudeResp.Tool != nil {
				toolInfo = &ToolInfo{
					ToolUseID: claudeResp.Tool.ToolUseID,
					ToolName:  claudeResp.Tool.ToolName,
					Input:     inputMap,
				}
			}
			responseChan <- MessageResponse{
				Type:       claudeResp.Type,
				Tool:       toolInfo,
				ToolOutput: claudeResp.ToolOutput,
				IsError:    claudeResp.IsError,
			}
		}
	}

	// Generate title after the response is complete (before closing channel)
	if pendingTitleGeneration {
		title, err := ch.claudeExecutor.GenerateTitleSummary(titleUserMessage, titleAssistantMessage)
		if err == nil && title != "" {
			ch.sessionManager.UpdateSessionTitle(currentSessionID, title)
			// Send session_title to client so they can update the sidebar
			responseChan <- MessageResponse{
				Type:      "session_title",
				SessionID: currentSessionID,
				Title:     title,
			}
		}
	}

	// Close the channel now that we're done
	close(responseChan)
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
			continue
		}

		if att.Type == "image" {
			claudePath := ch.getClaudePath(physicalPath)
			sb.WriteString(fmt.Sprintf("[Image: %s]\nPlease read and analyze this image file: %s\n\n", att.Filename, claudePath))
		} else {
			fileContent, err := ch.readFileContent(physicalPath)
			if err != nil {
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
// Container mode: maps /workspace/uploads/... to WORKSPACE_PATH/uploads/...
// Local dev mode: returns absolute local path
func (ch *ChatHandler) getClaudePath(localPath string) string {
	if ch.workspacePath != "" {
		// Container mode: map container path to host path for Claude CLI
		relPath, err := filepath.Rel(ch.uploadDir, localPath)
		if err != nil {
			return localPath
		}
		return filepath.Join(ch.workspacePath, "uploads", relPath)
	}

	// Local dev mode: use absolute local path
	absPath, err := filepath.Abs(localPath)
	if err != nil {
		return localPath
	}
	return absPath
}

// getPhysicalPath converts an API path to a physical file path
// API path format: /api/uploads/{filename}
func (ch *ChatHandler) getPhysicalPath(apiPath string) string {
	// Remove /api/uploads/ prefix
	prefix := "/api/uploads/"
	if !strings.HasPrefix(apiPath, prefix) {
		return ""
	}

	filename := strings.TrimPrefix(apiPath, prefix)
	return filepath.Join(ch.uploadDir, filename)
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
