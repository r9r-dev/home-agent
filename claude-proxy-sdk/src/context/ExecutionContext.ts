/**
 * ExecutionContext - Encapsulates execution state for a single Claude request
 *
 * Tracks active tool calls, accumulated inputs, and streaming state
 * to properly correlate SDK events during message processing.
 */

import type { ToolCallInfo, UsageInfo } from "../types.js";

export class ExecutionContext {
  /** Track active tool calls by content block index */
  private activeToolCalls = new Map<number, ToolCallInfo>();

  /** Track accumulated input JSON strings by tool index */
  private activeToolInputs = new Map<number, string>();

  /** Map tool_use_id to index for correlating results */
  private toolUseIdToIndex = new Map<string, number>();

  /** Track if we've received streaming content to avoid duplicates */
  private _hasReceivedStreamContent = false;

  /** Track processed message IDs to deduplicate usage */
  private processedMessageIds = new Set<string>();

  /** Accumulated usage from all turns */
  private accumulatedUsage: UsageInfo = {
    input_tokens: 0,
    output_tokens: 0,
    cache_creation_input_tokens: 0,
    cache_read_input_tokens: 0,
  };

  /**
   * Check if streaming content has been received
   */
  get hasReceivedStreamContent(): boolean {
    return this._hasReceivedStreamContent;
  }

  /**
   * Mark that streaming content has been received
   */
  markStreamContentReceived(): void {
    this._hasReceivedStreamContent = true;
  }

  /**
   * Reset context for a new message turn
   * Called when a new assistant message starts
   */
  resetForNewMessage(): void {
    this._hasReceivedStreamContent = false;
    this.activeToolCalls.clear();
    this.activeToolInputs.clear();
    this.toolUseIdToIndex.clear();
  }

  /**
   * Register a new tool call
   * @param index - Content block index from SDK
   * @param toolInfo - Tool call information
   */
  registerToolCall(index: number, toolInfo: ToolCallInfo): void {
    this.activeToolCalls.set(index, toolInfo);
    this.toolUseIdToIndex.set(toolInfo.tool_use_id, index);
  }

  /**
   * Get tool call info by content block index
   * @param index - Content block index
   * @returns Tool call info or undefined
   */
  getToolCall(index: number): ToolCallInfo | undefined {
    return this.activeToolCalls.get(index);
  }

  /**
   * Get tool call info by tool_use_id
   * @param toolUseId - Tool use identifier
   * @returns Tool call info or undefined
   */
  getToolCallByUseId(toolUseId: string): ToolCallInfo | undefined {
    const index = this.toolUseIdToIndex.get(toolUseId);
    if (index === undefined) return undefined;
    return this.activeToolCalls.get(index);
  }

  /**
   * Append input JSON delta to accumulated input
   * @param index - Content block index
   * @param delta - JSON delta string to append
   */
  appendToolInput(index: number, delta: string): void {
    const current = this.activeToolInputs.get(index) || "";
    this.activeToolInputs.set(index, current + delta);
  }

  /**
   * Get accumulated input for a tool call
   * @param index - Content block index
   * @returns Raw accumulated input string or undefined
   */
  getAccumulatedInputRaw(index: number): string | undefined {
    return this.activeToolInputs.get(index);
  }

  /**
   * Get and parse accumulated input for a tool call by tool_use_id
   * @param toolUseId - Tool use identifier
   * @returns Parsed input object or null if not found/invalid
   */
  getAccumulatedInput(toolUseId: string): Record<string, unknown> | null {
    const index = this.toolUseIdToIndex.get(toolUseId);
    if (index === undefined) return null;

    const inputStr = this.activeToolInputs.get(index);
    if (!inputStr) return null;

    try {
      return JSON.parse(inputStr);
    } catch {
      return null;
    }
  }

  /**
   * Get tool name by tool_use_id
   * @param toolUseId - Tool use identifier
   * @returns Tool name or empty string if not found
   */
  getToolName(toolUseId: string): string {
    const toolInfo = this.getToolCallByUseId(toolUseId);
    return toolInfo?.tool_name || "";
  }

  /**
   * Get index for a tool_use_id
   * @param toolUseId - Tool use identifier
   * @returns Index or undefined if not found
   */
  getIndexByToolUseId(toolUseId: string): number | undefined {
    return this.toolUseIdToIndex.get(toolUseId);
  }

  /**
   * Add usage from an assistant message (deduplicated by message ID)
   * @param messageId - Message ID from SDK
   * @param usage - Usage object from the message
   */
  addUsage(messageId: string, usage: { input_tokens?: number; output_tokens?: number; cache_creation_input_tokens?: number; cache_read_input_tokens?: number }): void {
    // Skip if already processed this message ID (handles parallel tool uses)
    if (this.processedMessageIds.has(messageId)) {
      return;
    }
    this.processedMessageIds.add(messageId);

    // Accumulate usage
    this.accumulatedUsage.input_tokens += usage.input_tokens || 0;
    this.accumulatedUsage.output_tokens += usage.output_tokens || 0;
    this.accumulatedUsage.cache_creation_input_tokens = (this.accumulatedUsage.cache_creation_input_tokens || 0) + (usage.cache_creation_input_tokens || 0);
    this.accumulatedUsage.cache_read_input_tokens = (this.accumulatedUsage.cache_read_input_tokens || 0) + (usage.cache_read_input_tokens || 0);

    console.log(`[Usage] Added from ${messageId}: in=${usage.input_tokens}, out=${usage.output_tokens}, cumulative: in=${this.accumulatedUsage.input_tokens}, out=${this.accumulatedUsage.output_tokens}`);
  }

  /**
   * Get accumulated usage from all turns
   * @returns Total accumulated usage
   */
  getAccumulatedUsage(): UsageInfo {
    return { ...this.accumulatedUsage };
  }

  /**
   * Check if we have any accumulated usage
   */
  hasUsage(): boolean {
    return this.accumulatedUsage.input_tokens > 0 || this.accumulatedUsage.output_tokens > 0;
  }
}
