package services

import (
	"context"
	"strings"
)

// ToolCallInfo represents information about a tool call
type ToolCallInfo struct {
	ToolUseID       string                 `json:"tool_use_id"`
	ToolName        string                 `json:"tool_name"`
	Input           map[string]interface{} `json:"input"`
	ParentToolUseID string                 `json:"parent_tool_use_id,omitempty"`
}

// UsageInfo represents token usage information
type UsageInfo struct {
	InputTokens              int     `json:"input_tokens"`
	OutputTokens             int     `json:"output_tokens"`
	CacheCreationInputTokens int     `json:"cache_creation_input_tokens,omitempty"`
	CacheReadInputTokens     int     `json:"cache_read_input_tokens,omitempty"`
	TotalCostUSD             float64 `json:"total_cost_usd,omitempty"`
}

// ClaudeResponse represents a chunk of text from Claude's response
type ClaudeResponse struct {
	Type      string // "chunk", "thinking", "done", "error", "session_id", "tool_start", "tool_progress", "tool_result", "tool_error", "tool_input_delta", "usage"
	Content   string
	SessionID string
	Error     error
	// Tool-specific fields
	Tool               *ToolCallInfo
	ElapsedTimeSeconds float64
	ToolOutput         string
	IsError            bool
	InputDelta         string // JSON delta for streaming tool input
	// Usage information
	Usage *UsageInfo
}

// ClaudeExecutor is the interface for executing Claude CLI commands
// via the Claude Proxy service.
type ClaudeExecutor interface {
	// ExecuteClaude executes Claude with the given prompt, session ID, model, and custom instructions.
	// sessionID: The session UUID to use (required for session management)
	// isNewSession: If true, uses --session-id to start a new session; if false, uses --resume
	// Model can be "haiku", "sonnet", or "opus" (defaults to "haiku" if empty).
	// customInstructions are appended to the system prompt if provided.
	// thinking: If true, enables extended thinking mode.
	// Returns a channel that streams ClaudeResponse events.
	// The channel will be closed when the execution completes.
	ExecuteClaude(ctx context.Context, prompt string, sessionID string, isNewSession bool, model string, customInstructions string, thinking bool) (<-chan ClaudeResponse, error)

	// GenerateTitleSummary generates a short title for a conversation.
	// Uses a fast model (haiku) for quick generation.
	GenerateTitleSummary(userMessage, assistantResponse string) (string, error)

	// TestConnection tests connectivity to the proxy service.
	TestConnection() error
}

// baseSystemPrompt is the default system prompt for Home Agent
const baseSystemPrompt = `You are a helpful personal assistant named Halfred, running on the user's home server.
You are here to help with ANY question or task the user might have.

## Your capabilities
You have access to various tools:
- **Command line**: Execute bash commands on the home server
- **Web search**: Search the internet for current information (weather, news, etc.)
- **Web fetch**: Retrieve content from web pages
- **File operations**: Read, write, and edit files

## What you can help with
- General questions (weather, facts, recommendations, etc.)
- Home server administration and monitoring
- Container management (Docker)
- System troubleshooting
- Network configuration
- Any other task the user requests

## Guidelines
- For questions requiring current information (weather, news, etc.), use the WebSearch tool
- Be careful with destructive commands and confirm before making significant changes
- Respond in the same language as the user
- Be concise but helpful`

// GetSystemPrompt returns the base system prompt for display in frontend
func GetSystemPrompt() string {
	return baseSystemPrompt
}

// BuildSystemPrompt constructs the final system prompt with optional custom instructions
func BuildSystemPrompt(customInstructions string) string {
	if customInstructions == "" {
		return baseSystemPrompt
	}
	return baseSystemPrompt + "\n\n## Instructions personnalisees\n" + customInstructions
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
