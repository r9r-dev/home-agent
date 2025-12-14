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

// ClaudeService handles interaction with Claude Code CLI
type ClaudeService struct {
	claudeBin string // Path to the claude binary
	timeout   time.Duration
}

// NewClaudeService creates a new ClaudeService instance
func NewClaudeService(claudeBin string) *ClaudeService {
	if claudeBin == "" {
		claudeBin = "claude" // Default to PATH lookup
	}

	log.Printf("Initializing ClaudeService with binary: %s", claudeBin)

	return &ClaudeService{
		claudeBin: claudeBin,
		timeout:   10 * time.Minute, // Default 10 minute timeout
	}
}

// SetTimeout sets the timeout for Claude command execution
func (cs *ClaudeService) SetTimeout(timeout time.Duration) {
	cs.timeout = timeout
}

// ExecuteClaude executes the Claude Code CLI and streams the response
// If sessionID is provided, it resumes the existing session
func (cs *ClaudeService) ExecuteClaude(ctx context.Context, prompt string, sessionID string) (<-chan ClaudeResponse, error) {
	log.Printf("Executing Claude with prompt (length: %d), sessionID: %s", len(prompt), sessionID)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, cs.timeout)

	// Build command arguments
	args := []string{"-p", prompt, "--output-format", "stream-json", "--verbose", "--model", "haiku"}

	// Add resume flag if session ID is provided
	if sessionID != "" {
		args = append(args, "--resume", sessionID)
		log.Printf("Resuming Claude session: %s", sessionID)
	}

	// Create the command
	cmd := exec.CommandContext(ctx, cs.claudeBin, args...)

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
	go cs.readStderr(stderr)

	// Start goroutine to process stdout
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

// processStream processes the JSON stream from Claude and sends responses to the channel
func (cs *ClaudeService) processStream(ctx context.Context, cancel context.CancelFunc, cmd *exec.Cmd, stdout io.ReadCloser, responseChan chan<- ClaudeResponse) {
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
