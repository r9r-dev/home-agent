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
	Type      string // "chunk", "thinking", "done", "error", "session_id"
	Content   string
	SessionID string
	Error     error
}

// ClaudeService handles direct interaction with Claude Code CLI
type ClaudeService struct {
	claudeBin string
	timeout   time.Duration
}

// NewClaudeService creates a new ClaudeService instance
func NewClaudeService(claudeBin string) *ClaudeService {
	if claudeBin == "" {
		claudeBin = "claude"
	}

	log.Printf("Initializing ClaudeService with binary: %s", claudeBin)

	return &ClaudeService{
		claudeBin: claudeBin,
		timeout:   10 * time.Minute,
	}
}

// SetTimeout sets the timeout for Claude command execution
func (cs *ClaudeService) SetTimeout(timeout time.Duration) {
	cs.timeout = timeout
}

// System prompt for Home Agent
const systemPrompt = `You are a system administrator assistant running on a home server infrastructure.
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

// ExecuteClaude executes the Claude Code CLI and streams the response
// sessionID: The session UUID to use for this conversation
// isNewSession: If true, uses --session-id to start a new session; if false, uses --resume
// model: Claude model (haiku, sonnet, opus) - defaults to sonnet if empty
// customInstructions: Optional custom instructions to append to system prompt
// thinking: If true, enables extended thinking mode with --thinking flag
func (cs *ClaudeService) ExecuteClaude(ctx context.Context, prompt string, sessionID string, isNewSession bool, model string, customInstructions string, thinking bool) (<-chan ClaudeResponse, error) {
	// Default to sonnet if model not specified
	if model == "" {
		model = "sonnet"
	}

	log.Printf("Executing Claude with prompt (length: %d), sessionID: %s, isNewSession: %v, model: %s, thinking: %v", len(prompt), sessionID, isNewSession, model, thinking)

	ctx, cancel := context.WithTimeout(ctx, cs.timeout)

	// Build the system prompt with custom instructions
	finalSystemPrompt := systemPrompt
	if customInstructions != "" {
		finalSystemPrompt = systemPrompt + "\n\n## Instructions personnalisees\n" + customInstructions
	}

	args := []string{
		"-p", prompt,
		"--output-format", "stream-json",
		"--verbose",
		"--model", model,
		"--system-prompt", finalSystemPrompt,
		"--dangerously-skip-permissions",
	}

	// Add thinking mode if enabled
	if thinking {
		args = append(args, "--thinking")
		log.Println("Thinking mode enabled")
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

	cmd := exec.CommandContext(ctx, cs.claudeBin, args...)
	cmd.Env = os.Environ()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start claude command: %w", err)
	}

	log.Printf("Claude command started (PID: %d)", cmd.Process.Pid)

	responseChan := make(chan ClaudeResponse, 100)

	go cs.readStderr(stderr)
	go cs.processStream(ctx, cancel, cmd, stdout, responseChan)

	return responseChan, nil
}

// readStderr reads and logs stderr output from Claude
func (cs *ClaudeService) readStderr(stderr io.ReadCloser) {
	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			log.Printf("Claude stderr: %s", line)
		}
	}
}

// processStream processes the JSON stream from Claude
func (cs *ClaudeService) processStream(ctx context.Context, cancel context.CancelFunc, cmd *exec.Cmd, stdout io.ReadCloser, responseChan chan<- ClaudeResponse) {
	defer close(responseChan)
	defer cancel()
	defer stdout.Close()

	scanner := bufio.NewScanner(stdout)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	var fullResponse strings.Builder
	var detectedSessionID string
	var hasContent bool // Track if we've already sent content (for paragraph separation)

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

		var event ClaudeStreamEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			log.Printf("Failed to parse JSON line: %v, line: %s", err, line)
			continue
		}

		switch event.Type {
		case "system":
			if event.SessionID != "" {
				detectedSessionID = event.SessionID
				log.Printf("Detected session ID from system event: %s", event.SessionID)
			}

		case "assistant":
			log.Println("Stream: assistant message")
			if event.Message != nil {
				if content, ok := event.Message["content"].([]interface{}); ok {
					for _, item := range content {
						if contentMap, ok := item.(map[string]interface{}); ok {
							if contentType, ok := contentMap["type"].(string); ok && contentType == "text" {
								if text, ok := contentMap["text"].(string); ok && text != "" {
									// Add paragraph separator if we already have content
									// This separates multiple assistant responses (e.g., before and after tool use)
									if hasContent {
										fullResponse.WriteString("\n\n")
										responseChan <- ClaudeResponse{
											Type:    "chunk",
											Content: "\n\n",
										}
									}
									fullResponse.WriteString(text)
									responseChan <- ClaudeResponse{
										Type:    "chunk",
										Content: text,
									}
									hasContent = true
								}
							}
						}
					}
				}
			}

		case "result":
			log.Println("Stream: result")
			if event.SessionID != "" {
				detectedSessionID = event.SessionID
			}

		case "message_start":
			log.Println("Stream: message_start")
			if event.Message != nil {
				if sid, ok := event.Message["session_id"].(string); ok && sid != "" {
					detectedSessionID = sid
					log.Printf("Detected session ID from message_start: %s", sid)
				}
			}

		case "content_block_start":
			log.Println("Stream: content_block_start")

		case "content_block_delta":
			if event.Delta != nil && event.Delta.Text != "" {
				if event.Delta.Type == "thinking_delta" {
					// Thinking content - send as separate type
					responseChan <- ClaudeResponse{
						Type:    "thinking",
						Content: event.Delta.Text,
					}
				} else if event.Delta.Type == "text_delta" {
					// Regular text content
					fullResponse.WriteString(event.Delta.Text)
					responseChan <- ClaudeResponse{
						Type:    "chunk",
						Content: event.Delta.Text,
					}
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

	if err := cmd.Wait(); err != nil {
		log.Printf("Command finished with error: %v", err)
		if fullResponse.Len() == 0 {
			responseChan <- ClaudeResponse{
				Type:  "error",
				Error: fmt.Errorf("claude command failed: %w", err),
			}
			return
		}
	}

	log.Printf("Stream processing complete. Total response length: %d", fullResponse.Len())

	if detectedSessionID != "" {
		responseChan <- ClaudeResponse{
			Type:      "session_id",
			SessionID: detectedSessionID,
		}
	}

	responseChan <- ClaudeResponse{
		Type:      "done",
		Content:   fullResponse.String(),
		SessionID: detectedSessionID,
	}
}

// GenerateTitleSummary generates a short title summary for a conversation
func (cs *ClaudeService) GenerateTitleSummary(userMessage, assistantResponse string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

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

	cmd := exec.CommandContext(ctx, cs.claudeBin,
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
	title = strings.Trim(title, "\"'")
	if len(title) > 50 {
		title = title[:47] + "..."
	}

	return title, nil
}

// TestClaudeBinary tests if the Claude binary is accessible
func (cs *ClaudeService) TestClaudeBinary() error {
	log.Printf("Testing Claude binary: %s", cs.claudeBin)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, cs.claudeBin, "--version")
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("failed to execute claude --version: %w (output: %s)", err, string(output))
	}

	log.Printf("Claude binary test successful: %s", strings.TrimSpace(string(output)))
	return nil
}
