package services

import (
	"context"
)

// ClaudeExecutor is the interface for executing Claude CLI commands.
// It can be implemented by either a local executor (direct CLI execution)
// or a proxy executor (remote execution via WebSocket).
type ClaudeExecutor interface {
	// ExecuteClaude executes Claude with the given prompt, optional session ID, and model.
	// Model can be "haiku", "sonnet", or "opus" (defaults to "haiku" if empty).
	// Returns a channel that streams ClaudeResponse events.
	// The channel will be closed when the execution completes.
	ExecuteClaude(ctx context.Context, prompt string, sessionID string, model string) (<-chan ClaudeResponse, error)

	// GenerateTitleSummary generates a short title for a conversation.
	// Uses a fast model (haiku) for quick generation.
	GenerateTitleSummary(userMessage, assistantResponse string) (string, error)

	// TestConnection tests if the executor is available and working.
	// For local executor, this tests the Claude binary.
	// For proxy executor, this tests connectivity to the proxy service.
	TestConnection() error
}
