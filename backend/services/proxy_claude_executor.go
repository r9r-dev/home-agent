package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// ProxyClaudeExecutor connects to a remote Claude proxy service via WebSocket.
// It implements the ClaudeExecutor interface for remote execution.
type ProxyClaudeExecutor struct {
	proxyURL string        // WebSocket URL, e.g., "ws://192.168.1.100:9090"
	httpURL  string        // HTTP URL, e.g., "http://192.168.1.100:9090"
	apiKey   string        // API key for authentication
	timeout  time.Duration // Timeout for operations
	dialer   *websocket.Dialer
}

// ProxyConfig holds configuration for the proxy executor
type ProxyConfig struct {
	ProxyURL string        // WebSocket URL for streaming (e.g., "ws://host:9090")
	APIKey   string        // Optional API key for authentication
	Timeout  time.Duration // Timeout for operations (default 10 minutes)
}

// ProxyRequest represents a request sent to the proxy
type ProxyRequest struct {
	Type               string `json:"type"`                          // "execute"
	Prompt             string `json:"prompt"`                        // The prompt to send to Claude
	SessionID          string `json:"session_id,omitempty"`          // Session ID (UUID)
	IsNewSession       bool   `json:"is_new_session,omitempty"`      // True for new session (--session-id), false for resume (--resume)
	Model              string `json:"model,omitempty"`               // Claude model: haiku, sonnet, opus
	CustomInstructions string `json:"custom_instructions,omitempty"` // Optional custom instructions for system prompt
	Thinking           bool   `json:"thinking,omitempty"`            // Enable extended thinking mode
}

// ProxyResponse represents a response from the proxy
type ProxyResponse struct {
	Type      string `json:"type"`                 // "chunk", "thinking", "done", "error", "session_id"
	Content   string `json:"content,omitempty"`    // Response content
	SessionID string `json:"session_id,omitempty"` // Session ID from Claude
	Error     string `json:"error,omitempty"`      // Error message
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

// NewProxyClaudeExecutor creates a new ProxyClaudeExecutor instance
func NewProxyClaudeExecutor(config ProxyConfig) *ProxyClaudeExecutor {
	if config.Timeout == 0 {
		config.Timeout = 10 * time.Minute
	}

	// Convert HTTP URL to WebSocket URL if needed
	wsURL := config.ProxyURL
	if strings.HasPrefix(wsURL, "http://") {
		wsURL = "ws://" + strings.TrimPrefix(wsURL, "http://")
	} else if strings.HasPrefix(wsURL, "https://") {
		wsURL = "wss://" + strings.TrimPrefix(wsURL, "https://")
	}

	// Convert WebSocket URL to HTTP URL
	httpURL := config.ProxyURL
	if strings.HasPrefix(httpURL, "ws://") {
		httpURL = "http://" + strings.TrimPrefix(httpURL, "ws://")
	} else if strings.HasPrefix(httpURL, "wss://") {
		httpURL = "https://" + strings.TrimPrefix(httpURL, "wss://")
	}

	return &ProxyClaudeExecutor{
		proxyURL: wsURL,
		httpURL:  httpURL,
		apiKey:   config.APIKey,
		timeout:  config.Timeout,
		dialer: &websocket.Dialer{
			HandshakeTimeout: 10 * time.Second,
		},
	}
}

// ExecuteClaude connects to the proxy service and streams Claude's response
func (pce *ProxyClaudeExecutor) ExecuteClaude(ctx context.Context, prompt string, sessionID string, isNewSession bool, model string, customInstructions string, thinking bool) (<-chan ClaudeResponse, error) {
	// Default to haiku if model not specified
	if model == "" {
		model = "haiku"
	}

	responseChan := make(chan ClaudeResponse, 100)

	go func() {
		defer close(responseChan)

		// Create context with timeout
		ctx, cancel := context.WithTimeout(ctx, pce.timeout)
		defer cancel()

		// Connect to proxy with retries
		conn, err := pce.connectWithRetry(ctx, 3)
		if err != nil {
			log.Printf("ProxyExecutor: Failed to connect to proxy: %v", err)
			responseChan <- ClaudeResponse{
				Type:  "error",
				Error: fmt.Errorf("failed to connect to proxy: %w", err),
			}
			return
		}
		defer conn.Close()

		// Send execute request
		request := ProxyRequest{
			Type:               "execute",
			Prompt:             prompt,
			SessionID:          sessionID,
			IsNewSession:       isNewSession,
			Model:              model,
			CustomInstructions: customInstructions,
			Thinking:           thinking,
		}

		if err := conn.WriteJSON(request); err != nil {
			log.Printf("ProxyExecutor: Failed to send request: %v", err)
			responseChan <- ClaudeResponse{
				Type:  "error",
				Error: fmt.Errorf("failed to send request: %w", err),
			}
			return
		}

		// Read responses
		for {
			select {
			case <-ctx.Done():
				responseChan <- ClaudeResponse{
					Type:  "error",
					Error: fmt.Errorf("request cancelled"),
				}
				return
			default:
			}

			var response ProxyResponse
			if err := conn.ReadJSON(&response); err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					return
				}
				responseChan <- ClaudeResponse{
					Type:  "error",
					Error: fmt.Errorf("failed to read response: %w", err),
				}
				return
			}

			// Convert proxy response to ClaudeResponse
			switch response.Type {
			case "chunk":
				responseChan <- ClaudeResponse{
					Type:    "chunk",
					Content: response.Content,
				}

			case "thinking":
				responseChan <- ClaudeResponse{
					Type:    "thinking",
					Content: response.Content,
				}

			case "session_id":
				responseChan <- ClaudeResponse{
					Type:      "session_id",
					SessionID: response.SessionID,
				}

			case "done":
				responseChan <- ClaudeResponse{
					Type:      "done",
					Content:   response.Content,
					SessionID: response.SessionID,
				}
				return

			case "error":
				responseChan <- ClaudeResponse{
					Type:  "error",
					Error: fmt.Errorf("proxy error: %s", response.Error),
				}
				return
			}
		}
	}()

	return responseChan, nil
}

// connectWithRetry attempts to connect to the proxy with retries
func (pce *ProxyClaudeExecutor) connectWithRetry(ctx context.Context, maxAttempts int) (*websocket.Conn, error) {
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		conn, err := pce.dial(ctx)
		if err == nil {
			return conn, nil
		}

		lastErr = err

		if attempt < maxAttempts {
			// Exponential backoff
			backoff := time.Duration(attempt) * time.Second
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}
	}

	return nil, fmt.Errorf("failed to connect after %d attempts: %w", maxAttempts, lastErr)
}

// dial establishes a WebSocket connection to the proxy
func (pce *ProxyClaudeExecutor) dial(ctx context.Context) (*websocket.Conn, error) {
	wsURL := pce.proxyURL + "/ws"

	header := http.Header{}
	if pce.apiKey != "" {
		header.Set("X-API-Key", pce.apiKey)
	}

	conn, _, err := pce.dialer.DialContext(ctx, wsURL, header)
	if err != nil {
		return nil, fmt.Errorf("failed to dial %s: %w", wsURL, err)
	}

	return conn, nil
}

// GenerateTitleSummary generates a title via HTTP request to the proxy
func (pce *ProxyClaudeExecutor) GenerateTitleSummary(userMessage, assistantResponse string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Truncate messages if too long
	if len(userMessage) > 500 {
		userMessage = userMessage[:500]
	}
	if len(assistantResponse) > 500 {
		assistantResponse = assistantResponse[:500]
	}

	request := TitleRequest{
		UserMessage:       userMessage,
		AssistantResponse: assistantResponse,
	}

	body, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := pce.httpURL + "/api/title"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if pce.apiKey != "" {
		req.Header.Set("X-API-Key", pce.apiKey)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("proxy returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var titleResp TitleResponse
	if err := json.NewDecoder(resp.Body).Decode(&titleResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if titleResp.Error != "" {
		return "", fmt.Errorf("proxy error: %s", titleResp.Error)
	}

	return titleResp.Title, nil
}

// TestConnection tests connectivity to the proxy service
func (pce *ProxyClaudeExecutor) TestConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	url := pce.httpURL + "/health"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if pce.apiKey != "" {
		req.Header.Set("X-API-Key", pce.apiKey)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to proxy at %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("proxy health check failed with status %d", resp.StatusCode)
	}

	return nil
}
