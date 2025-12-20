/**
 * Claude Agent SDK wrapper
 * Provides streaming execution with hooks support
 */

import { query, type Options, type SDKMessage, type PreToolUseHookInput, type PostToolUseHookInput } from "@anthropic-ai/claude-agent-sdk";
import type { ProxyRequest, ProxyResponse, ToolCallInfo } from "./types.js";
import { auditLog } from "./hooks/audit.js";

// Execution context for tracking tool calls within a single execution
interface ExecutionContext {
  // Track active tool calls by index (for correlating content blocks)
  activeToolCalls: Map<number, ToolCallInfo>;
  // Track accumulated input JSON strings by tool index
  activeToolInputs: Map<number, string>;
  // Map tool_use_id to index for correlating results
  toolUseIdToIndex: Map<string, number>;
  // Track if we've received streaming content to avoid duplicates
  hasReceivedStreamContent: boolean;
}

// System prompt for Home Agent
const SYSTEM_PROMPT = `You are a helpful personal assistant named Halfred, running on the user's home server.
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
- Be concise but helpful`;

// Model aliases supported by Claude Agent SDK
function mapModel(model?: string): string {
  // SDK supports simple aliases: "haiku", "sonnet", "opus"
  // These always point to the latest version of each model
  switch (model) {
    case "haiku":
    case "opus":
      return model;
    case "sonnet":
    default:
      return "sonnet";
  }
}

/**
 * Execute a prompt using Claude Agent SDK
 * Yields ProxyResponse objects compatible with the Go backend protocol
 */
export async function* executePrompt(
  request: ProxyRequest
): AsyncGenerator<ProxyResponse> {
  const {
    prompt,
    session_id,
    is_new_session,
    model,
    custom_instructions,
    thinking,
  } = request;

  // Create execution context local to this request
  const ctx: ExecutionContext = {
    activeToolCalls: new Map(),
    activeToolInputs: new Map(),
    toolUseIdToIndex: new Map(),
    hasReceivedStreamContent: false,
  };

  // Build system prompt with custom instructions
  let systemPrompt = SYSTEM_PROMPT;
  if (custom_instructions) {
    systemPrompt += `\n\n## Instructions personnalisees\n${custom_instructions}`;
  }

  // Build Agent SDK options
  const options: Options = {
    // Tools available to the agent
    tools: ["Read", "Write", "Edit", "Bash", "Glob", "Grep", "WebSearch", "WebFetch"],

    // Auto-allow these tools
    allowedTools: ["Read", "Write", "Edit", "Bash", "Glob", "Grep", "WebSearch", "WebFetch"],

    // Permission mode - accept edits without prompting
    permissionMode: "acceptEdits",

    // Model selection
    model: mapModel(model),

    // System prompt
    systemPrompt,

    // Include streaming events
    includePartialMessages: true,

    // Extended thinking mode
    ...(thinking && {
      maxThinkingTokens: 10000,
    }),

    // Session management
    ...(session_id &&
      !is_new_session && {
        resume: session_id,
      }),

    // Hooks for auditing (using callback hooks instead of shell commands)
    hooks: {
      PreToolUse: [
        {
          matcher: "Bash",
          hooks: [
            async (input, _toolUseId, _options) => {
              const hookInput = input as PreToolUseHookInput;
              auditLog({
                timestamp: new Date(),
                sessionId: session_id,
                event: "PreToolUse",
                tool: "Bash",
                details: { input: hookInput.tool_input },
              });
              return { continue: true };
            },
          ],
        },
      ],
      PostToolUse: [
        {
          matcher: "Edit|Write",
          hooks: [
            async (input, _toolUseId, _options) => {
              const hookInput = input as PostToolUseHookInput;
              auditLog({
                timestamp: new Date(),
                sessionId: session_id,
                event: "PostToolUse",
                tool: hookInput.tool_name,
                details: { input: hookInput.tool_input },
              });
              return { continue: true };
            },
          ],
        },
      ],
    },
  };

  auditLog({
    timestamp: new Date(),
    sessionId: session_id,
    event: "execute",
    details: { model, thinking, is_new_session },
  });

  let detectedSessionId: string | undefined;
  let fullResponse = "";

  try {
    for await (const message of query({ prompt, options })) {
      // Debug: log ALL raw SDK messages to understand the protocol
      console.log(`[SDK RAW] ${JSON.stringify(message).substring(0, 500)}`);

      const response = processMessage(message, ctx, session_id);

      if (response) {
        // Debug: log response being sent
        console.log(`[Proxy] Sending: ${response.type}`, response.tool ? `tool: ${response.tool.tool_name} (${response.tool.tool_use_id})` : "");

        // Track session ID
        if (response.type === "session_id" && response.session_id) {
          detectedSessionId = response.session_id;
        }

        // Accumulate text content
        if (response.type === "chunk" && response.content) {
          fullResponse += response.content;
        }

        yield response;
      }
    }

    // Send done message
    yield {
      type: "done",
      content: fullResponse,
      session_id: detectedSessionId || session_id,
    };
  } catch (error) {
    const errorMessage =
      error instanceof Error ? error.message : "Unknown error";
    auditLog({
      timestamp: new Date(),
      sessionId: session_id,
      event: "error",
      details: { error: errorMessage },
    });

    yield {
      type: "error",
      error: errorMessage,
    };
  }
}

/**
 * Process an SDK message and convert to ProxyResponse
 * Uses execution context to track state for this specific execution
 */
function processMessage(message: SDKMessage, ctx: ExecutionContext, sessionId?: string): ProxyResponse | null {
  switch (message.type) {
    case "system":
      if (message.subtype === "init") {
        // Reset streaming flag for new session (context is already fresh)
        ctx.hasReceivedStreamContent = false;
        // Capture session ID from init message
        return {
          type: "session_id",
          session_id: message.session_id,
        };
      }
      break;

    case "assistant":
      // Skip full assistant message if we already streamed the content
      // This avoids duplicating the response
      if (ctx.hasReceivedStreamContent) {
        break;
      }
      // Full assistant message (non-streaming fallback)
      if (message.message?.content) {
        const content = message.message.content;
        if (Array.isArray(content)) {
          for (const block of content) {
            if (block.type === "text" && "text" in block) {
              return {
                type: "chunk",
                content: block.text,
              };
            } else if (block.type === "thinking" && "thinking" in block) {
              return {
                type: "thinking",
                content: (block as { type: "thinking"; thinking: string }).thinking,
              };
            }
          }
        }
      }
      break;

    case "stream_event":
      // Streaming delta content
      const event = message.event;

      // Reset streaming flag on message_start - this indicates a new turn
      // This allows subsequent thinking blocks to be captured after tool calls
      if (event.type === "message_start") {
        ctx.hasReceivedStreamContent = false;
      }

      // Handle content_block_start for tool_use
      if (event.type === "content_block_start") {
        const contentBlock = (event as { type: "content_block_start"; index: number; content_block: { type: string; id?: string; name?: string; input?: Record<string, unknown> } }).content_block;
        if (contentBlock?.type === "tool_use" && contentBlock.id && contentBlock.name) {
          const toolInfo: ToolCallInfo = {
            tool_use_id: contentBlock.id,
            tool_name: contentBlock.name,
            input: contentBlock.input || {},
          };
          // Store for later correlation
          const index = (event as { index: number }).index;
          ctx.activeToolCalls.set(index, toolInfo);
          ctx.toolUseIdToIndex.set(contentBlock.id, index);

          return {
            type: "tool_start",
            tool: toolInfo,
          };
        }
      }

      // Handle content_block_delta
      if (event.type === "content_block_delta") {
        const eventWithIndex = event as { type: "content_block_delta"; index: number; delta: { type: string; text?: string; thinking?: string; partial_json?: string } };
        const delta = eventWithIndex.delta;
        const blockIndex = eventWithIndex.index;

        if (delta.type === "text_delta" && "text" in delta) {
          ctx.hasReceivedStreamContent = true;
          return {
            type: "chunk",
            content: delta.text,
          };
        } else if (delta.type === "thinking_delta" && "thinking" in delta) {
          ctx.hasReceivedStreamContent = true;
          return {
            type: "thinking",
            content: delta.thinking as string,
          };
        } else if (delta.type === "input_json_delta" && "partial_json" in delta) {
          // Accumulate input JSON delta
          const partialJson = delta.partial_json as string;
          const currentInput = ctx.activeToolInputs.get(blockIndex) || "";
          ctx.activeToolInputs.set(blockIndex, currentInput + partialJson);

          // Get the tool info for this block
          const toolInfo = ctx.activeToolCalls.get(blockIndex);
          if (toolInfo) {
            return {
              type: "tool_input_delta",
              tool: {
                tool_use_id: toolInfo.tool_use_id,
                tool_name: toolInfo.tool_name,
                input: {},
              },
              input_delta: partialJson,
            };
          }
        }
      }

      // Handle content_block_stop - tool input is complete
      if (event.type === "content_block_stop") {
        const stopEvent = event as { type: "content_block_stop"; index: number };
        const toolInfo = ctx.activeToolCalls.get(stopEvent.index);
        if (toolInfo) {
          // Parse the accumulated input
          const inputJsonStr = ctx.activeToolInputs.get(stopEvent.index);
          if (inputJsonStr) {
            try {
              const parsedInput = JSON.parse(inputJsonStr);
              console.log(`[SDK] content_block_stop: tool ${toolInfo.tool_name} input complete:`, JSON.stringify(parsedInput).substring(0, 200));
            } catch (e) {
              console.log(`[SDK] content_block_stop: tool ${toolInfo.tool_name} input parse error`);
            }
          }
        }
      }
      break;

    case "tool_progress":
      // Tool execution progress update
      const toolProgress = message as { type: "tool_progress"; tool_use_id: string; tool_name: string; parent_tool_use_id: string | null; elapsed_time_seconds: number };
      return {
        type: "tool_progress",
        tool: {
          tool_use_id: toolProgress.tool_use_id,
          tool_name: toolProgress.tool_name,
          input: {},
          parent_tool_use_id: toolProgress.parent_tool_use_id,
        },
        elapsed_time_seconds: toolProgress.elapsed_time_seconds,
      };

    case "user":
      // Check for tool results in user messages
      // Tool results can be identified by:
      // 1. parent_tool_use_id at message level (some SDK versions)
      // 2. tool_result blocks in message.content (current SDK format)
      const userMessage = message as {
        type: "user";
        parent_tool_use_id?: string | null;
        message?: {
          content?: Array<{
            type: string;
            tool_use_id?: string;
            content?: string | Array<{ type: string; text?: string }>;
            is_error?: boolean;
          }>;
        };
      };

      // Look for tool_result blocks in message content
      if (userMessage.message?.content) {
        for (const block of userMessage.message.content) {
          if (block.type === "tool_result" && block.tool_use_id) {
            const toolUseId = block.tool_use_id;
            const isError = block.is_error || false;
            let toolOutput = "";

            if (typeof block.content === "string") {
              toolOutput = block.content;
            } else if (Array.isArray(block.content)) {
              // Content can be an array of text blocks
              toolOutput = block.content
                .filter((c) => c.type === "text" && c.text)
                .map((c) => c.text)
                .join("\n");
            }

            // Get accumulated input for this tool call
            const toolIndex = ctx.toolUseIdToIndex.get(toolUseId);
            let accumulatedInput: Record<string, unknown> = {};
            let toolName = "";

            if (toolIndex !== undefined) {
              // Parse accumulated input JSON
              const inputJsonStr = ctx.activeToolInputs.get(toolIndex);
              if (inputJsonStr) {
                try {
                  accumulatedInput = JSON.parse(inputJsonStr);
                } catch {
                  // If parsing fails, leave as empty object
                }
              }
              // Get tool name from stored info
              const toolInfo = ctx.activeToolCalls.get(toolIndex);
              if (toolInfo) {
                toolName = toolInfo.tool_name;
              }
            }

            console.log(`[SDK] Tool result for ${toolName} (${toolUseId}): error=${isError}, output_len=${toolOutput.length}`);

            return {
              type: isError ? "tool_error" : "tool_result",
              tool: {
                tool_use_id: toolUseId,
                tool_name: toolName,
                input: accumulatedInput,
              },
              tool_output: toolOutput,
              is_error: isError,
            };
          }
        }
      }
      break;

    case "result":
      // Final result - don't send as we handle this separately
      break;
  }

  return null;
}

/**
 * Generate a title summary for a conversation
 */
export async function generateTitle(
  userMessage: string,
  assistantResponse: string
): Promise<string> {
  const truncatedUser = userMessage.slice(0, 500);
  const truncatedAssistant = assistantResponse.slice(0, 500);

  const prompt = `Tu dois generer un titre EN FRANCAIS, tres court (maximum 40 caracteres) qui resume cette conversation.
IMPORTANT: Le titre doit etre en francais.
Reponds UNIQUEMENT avec le titre, sans guillemets, sans ponctuation finale, sans explication.

Message de l'utilisateur: ${truncatedUser}

Reponse de l'assistant: ${truncatedAssistant}`;

  const options: Options = {
    model: "haiku",
    tools: [],
    permissionMode: "bypassPermissions",
    allowDangerouslySkipPermissions: true,
    maxTurns: 1,
    includePartialMessages: true,
  };

  let title = "";
  let hasStreamedContent = false;

  for await (const message of query({ prompt, options })) {
    if (message.type === "stream_event") {
      const event = message.event;
      if (event.type === "content_block_delta" && event.delta.type === "text_delta" && "text" in event.delta) {
        hasStreamedContent = true;
        title += event.delta.text;
      }
    } else if (message.type === "assistant" && message.message?.content && !hasStreamedContent) {
      // Only use assistant message if we didn't receive streamed content
      const content = message.message.content;
      if (Array.isArray(content)) {
        for (const block of content) {
          if (block.type === "text" && "text" in block) {
            title += block.text;
          }
        }
      }
    }
  }

  // Clean up title
  title = title.trim().replace(/^["']|["']$/g, "");
  if (title.length > 50) {
    title = title.slice(0, 47) + "...";
  }

  return title;
}
