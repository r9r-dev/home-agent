package services

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

// ClaudeStreamEvent represents a JSON event from Claude Code's stream output
type ClaudeStreamEvent struct {
	Type         string                 `json:"type"`
	Message      map[string]interface{} `json:"message,omitempty"`
	Delta        *ClaudeDelta           `json:"delta,omitempty"`
	ContentBlock *ClaudeContentBlock    `json:"content_block,omitempty"`
	Index        *int                   `json:"index,omitempty"`
	SessionID    string                 `json:"session_id,omitempty"`
}

// ClaudeDelta represents the delta content in a stream event
type ClaudeDelta struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// ClaudeContentBlock represents a content block in a stream event
type ClaudeContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// ClaudeResponse represents a chunk of text from Claude's response
type ClaudeResponse struct {
	Type      string // "chunk", "done", "error", "session_id"
	Content   string
	SessionID string
	Error     error
}

// LocalClaudeExecutor handles direct interaction with Claude Code CLI.
// It implements the ClaudeExecutor interface for local execution.
type LocalClaudeExecutor struct {
	claudeBin string // Path to the claude binary
	timeout   time.Duration
}

// NewLocalClaudeExecutor creates a new LocalClaudeExecutor instance
func NewLocalClaudeExecutor(claudeBin string) *LocalClaudeExecutor {
	if claudeBin == "" {
		claudeBin = "claude" // Default to PATH lookup
	}

	log.Printf("Initializing LocalClaudeExecutor with binary: %s", claudeBin)

	return &LocalClaudeExecutor{
		claudeBin: claudeBin,
		timeout:   10 * time.Minute, // Default 10 minute timeout
	}
}

// SetTimeout sets the timeout for Claude command execution
func (lce *LocalClaudeExecutor) SetTimeout(timeout time.Duration) {
	lce.timeout = timeout
}

// baseSystemPrompt is the default system prompt for Home Agent
const baseSystemPrompt = `You are a system administrator assistant running on a home server infrastructure.
You have access to the command line and can execute commands to help manage and monitor the systems.
Your role is to help with:
- Server administration and maintenance
- Container management (Docker)
- System monitoring and troubleshooting
- Network configuration
- Security audits and hardening
- Backup and recovery operations

You are NOT in a development environment. You are managing production home infrastructure.
Be careful with destructive commands and always confirm before making significant changes.
Respond in the same language as the user.`

// GetSystemPrompt returns the base system prompt for display in frontend
func GetSystemPrompt() string {
	return baseSystemPrompt
}

// MemoryEntry represents a memory item for prompt injection
type MemoryEntry struct {
	Title   string
	Content string
}

// FormatMemoryEntries formats memory entries for injection into the prompt
func FormatMemoryEntries(entries []MemoryEntry) string {
	if len(entries) == 0 {
		return ""
	}

	var builder strings.Builder
	builder.WriteString("<user_memory>\n")
	for _, entry := range entries {
		builder.WriteString("- ")
		builder.WriteString(entry.Title)
		builder.WriteString(": ")
		builder.WriteString(entry.Content)
		builder.WriteString("\n")
	}
	builder.WriteString("</user_memory>")
	return builder.String()
}

// buildSystemPrompt constructs the final system prompt with optional custom instructions
func buildSystemPrompt(customInstructions string) string {
	if customInstructions == "" {
		return baseSystemPrompt
	}
	return baseSystemPrompt + "\n\n## Instructions personnalisees\n" + customInstructions
}

// ExecuteClaude executes the Claude Code CLI and streams the response
// sessionID: The session UUID to use for this conversation
// isNewSession: If true, uses --session-id to start a new session; if false, uses --resume
// Model can be "haiku", "sonnet", or "opus" (defaults to "haiku" if empty)
// customInstructions are appended to the system prompt if provided
func (lce *LocalClaudeExecutor) ExecuteClaude(ctx context.Context, prompt string, sessionID string, isNewSession bool, model string, customInstructions string) (<-chan ClaudeResponse, error) {
	// Default to haiku if model not specified
	if model == "" {
		model = "haiku"
	}

	log.Printf("Executing Claude with prompt (length: %d), sessionID: %s, isNewSession: %v, model: %s, customInstructions: %d chars", len(prompt), sessionID, isNewSession, model, len(customInstructions))

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, lce.timeout)

	// Build the system prompt with custom instructions
	finalSystemPrompt := buildSystemPrompt(customInstructions)

	// Build command arguments
	args := []string{
		"-p", prompt,
		"--output-format", "stream-json",
		"--verbose",
		"--model", model,
		"--system-prompt", finalSystemPrompt,
		"--dangerously-skip-permissions",
	}

	// Add session management flag based on whether this is a new or existing session
	if sessionID != "" {
		if isNewSession {
			// New session: use --session-id to tell Claude to use our UUID
			args = append(args, "--session-id", sessionID)
			log.Printf("Starting new Claude session with ID: %s", sessionID)
		} else {
			// Existing session: use --resume to continue the conversation
			args = append(args, "--resume", sessionID)
			log.Printf("Resuming Claude session: %s", sessionID)
		}
	}

	// Create the command
	cmd := exec.CommandContext(ctx, lce.claudeBin, args...)

	// Set environment variables
	cmd.Env = os.Environ()

	// Get stdout pipe
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	// Get stderr pipe for error logging
	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start claude command: %w", err)
	}

	log.Printf("Claude command started (PID: %d)", cmd.Process.Pid)

	// Create response channel
	responseChan := make(chan ClaudeResponse, 100)

	// Start goroutine to read stderr
	go lce.readStderr(stderr)

	// Start goroutine to process stdout
	go lce.processStream(ctx, cancel, cmd, stdout, responseChan)

	return responseChan, nil
}

// readStderr reads and logs stderr output from Claude
func (lce *LocalClaudeExecutor) readStderr(stderr io.ReadCloser) {
	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			log.Printf("Claude stderr: %s", line)
		}
	}
}

// processStream processes the JSON stream from Claude and sends responses to the channel
func (lce *LocalClaudeExecutor) processStream(ctx context.Context, cancel context.CancelFunc, cmd *exec.Cmd, stdout io.ReadCloser, responseChan chan<- ClaudeResponse) {
	defer close(responseChan)
	defer cancel()
	defer stdout.Close()

	scanner := bufio.NewScanner(stdout)
	// Increase buffer size for large JSON lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	var fullResponse strings.Builder
	var detectedSessionID string

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			log.Println("Context cancelled, stopping stream processing")
			responseChan <- ClaudeResponse{
				Type:  "error",
				Error: fmt.Errorf("request cancelled"),
			}
			return
		default:
		}

		line := scanner.Text()
		if line == "" {
			continue
		}

		// Parse JSON event
		var event ClaudeStreamEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			log.Printf("Failed to parse JSON line: %v, line: %s", err, line)
			continue
		}

		// Handle different event types
		switch event.Type {
		case "system":
			// System event from verbose mode, contains session_id
			if event.SessionID != "" {
				detectedSessionID = event.SessionID
				log.Printf("Detected session ID from system event: %s", event.SessionID)
			}

		case "assistant":
			// Assistant message from verbose mode
			log.Println("Stream: assistant message")
			// Extract text content from message.content array
			if event.Message != nil {
				if content, ok := event.Message["content"].([]interface{}); ok {
					for _, item := range content {
						if contentMap, ok := item.(map[string]interface{}); ok {
							if contentType, ok := contentMap["type"].(string); ok && contentType == "text" {
								if text, ok := contentMap["text"].(string); ok && text != "" {
									fullResponse.WriteString(text)
									// Send the text as a single chunk
									responseChan <- ClaudeResponse{
										Type:    "chunk",
										Content: text,
									}
								}
							}
						}
					}
				}
			}

		case "result":
			// Result event from verbose mode - final event
			log.Println("Stream: result")
			if event.SessionID != "" {
				detectedSessionID = event.SessionID
			}

		case "message_start":
			log.Println("Stream: message_start")
			// Check if session_id is in the message
			if event.Message != nil {
				if sid, ok := event.Message["session_id"].(string); ok && sid != "" {
					detectedSessionID = sid
					log.Printf("Detected session ID from message_start: %s", sid)
				}
			}

		case "content_block_start":
			log.Println("Stream: content_block_start")

		case "content_block_delta":
			if event.Delta != nil && event.Delta.Type == "text_delta" && event.Delta.Text != "" {
				fullResponse.WriteString(event.Delta.Text)
				responseChan <- ClaudeResponse{
					Type:    "chunk",
					Content: event.Delta.Text,
				}
			}

		case "content_block_stop":
			log.Println("Stream: content_block_stop")

		case "message_delta":
			log.Println("Stream: message_delta")

		case "message_stop":
			log.Println("Stream: message_stop")

		case "error":
			log.Printf("Stream error event: %+v", event)
			responseChan <- ClaudeResponse{
				Type:  "error",
				Error: fmt.Errorf("claude error: %v", event.Message),
			}
			return

		default:
			// Check if this line contains a session_id at the root level
			if event.SessionID != "" {
				detectedSessionID = event.SessionID
				log.Printf("Detected session ID from event: %s", event.SessionID)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Scanner error: %v", err)
		responseChan <- ClaudeResponse{
			Type:  "error",
			Error: fmt.Errorf("scanner error: %w", err),
		}
		return
	}

	// Wait for command to finish
	if err := cmd.Wait(); err != nil {
		log.Printf("Command finished with error: %v", err)
		// Don't send error if we already got content - Claude might exit with non-zero on some conditions
		if fullResponse.Len() == 0 {
			responseChan <- ClaudeResponse{
				Type:  "error",
				Error: fmt.Errorf("claude command failed: %w", err),
			}
			return
		}
	}

	log.Printf("Stream processing complete. Total response length: %d", fullResponse.Len())

	// Send session ID if detected
	if detectedSessionID != "" {
		responseChan <- ClaudeResponse{
			Type:      "session_id",
			SessionID: detectedSessionID,
		}
	}

	// Send done signal with full response
	responseChan <- ClaudeResponse{
		Type:      "done",
		Content:   fullResponse.String(),
		SessionID: detectedSessionID,
	}
}

// GenerateTitleSummary generates a short title summary for a conversation using Claude
func (lce *LocalClaudeExecutor) GenerateTitleSummary(userMessage, assistantResponse string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Truncate messages if too long
	if len(userMessage) > 500 {
		userMessage = userMessage[:500]
	}
	if len(assistantResponse) > 500 {
		assistantResponse = assistantResponse[:500]
	}

	prompt := "Tu dois generer un titre EN FRANCAIS, tres court (maximum 40 caracteres) qui resume cette conversation. " +
		"IMPORTANT: Le titre doit etre en francais. " +
		"Reponds UNIQUEMENT avec le titre, sans guillemets, sans ponctuation finale, sans explication.\n\n" +
		"Message de l'utilisateur: " + userMessage + "\n\n" +
		"Reponse de l'assistant: " + assistantResponse

	// Use haiku model for quick title generation
	cmd := exec.CommandContext(ctx, lce.claudeBin,
		"-p", prompt,
		"--model", "haiku",
		"--max-turns", "1",
	)
	cmd.Env = os.Environ()

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to generate title: %w", err)
	}

	title := strings.TrimSpace(string(output))
	// Remove quotes if present
	title = strings.Trim(title, "\"'")
	// Truncate if too long
	if len(title) > 50 {
		title = title[:47] + "..."
	}

	return title, nil
}

// TestConnection tests if the Claude binary is accessible.
// Implements ClaudeExecutor interface.
func (lce *LocalClaudeExecutor) TestConnection() error {
	log.Printf("Testing Claude binary: %s", lce.claudeBin)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, lce.claudeBin, "--version")
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("failed to execute claude --version: %w (output: %s)", err, string(output))
	}

	log.Printf("Claude binary test successful: %s", strings.TrimSpace(string(output)))
	return nil
}
