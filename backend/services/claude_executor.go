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

// ClaudeResponse represents a chunk of text from Claude's response
type ClaudeResponse struct {
	Type      string // "chunk", "thinking", "done", "error", "session_id", "tool_start", "tool_progress", "tool_result", "tool_error"
	Content   string
	SessionID string
	Error     error
	// Tool-specific fields
	Tool               *ToolCallInfo
	ElapsedTimeSeconds float64
	ToolOutput         string
	IsError            bool
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
