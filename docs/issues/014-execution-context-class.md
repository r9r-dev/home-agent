# Extract ExecutionContext as class

**Priority:** P3 (Low)
**Type:** Refactoring
**Component:** Claude Proxy SDK
**Estimated Effort:** Low

## Summary

Convert `ExecutionContext` from inline interface to a proper class with encapsulated methods.

## Current State

Context is defined as an interface with direct property access:

```typescript
interface ExecutionContext {
  activeToolCalls: Map<number, ToolCallInfo>;
  activeToolInputs: Map<number, string>;
  toolUseIdToIndex: Map<string, number>;
  hasReceivedStreamContent: boolean;
}

// Direct manipulation throughout code
ctx.activeToolCalls.set(index, toolInfo);
ctx.toolUseIdToIndex.set(toolInfo.tool_use_id, index);
const current = ctx.activeToolInputs.get(index) || '';
ctx.activeToolInputs.set(index, current + delta);
```

## Proposed Solution

```typescript
// src/context/ExecutionContext.ts
export class ExecutionContext {
  private activeToolCalls = new Map<number, ToolCallInfo>();
  private activeToolInputs = new Map<number, string>();
  private toolUseIdToIndex = new Map<string, number>();
  private _hasReceivedStreamContent = false;

  get hasReceivedStreamContent(): boolean {
    return this._hasReceivedStreamContent;
  }

  markStreamContentReceived(): void {
    this._hasReceivedStreamContent = true;
  }

  resetForNewMessage(): void {
    this._hasReceivedStreamContent = false;
    this.activeToolCalls.clear();
    this.activeToolInputs.clear();
    this.toolUseIdToIndex.clear();
  }

  registerToolCall(index: number, toolInfo: ToolCallInfo): void {
    this.activeToolCalls.set(index, toolInfo);
    this.toolUseIdToIndex.set(toolInfo.tool_use_id, index);
  }

  getToolCall(index: number): ToolCallInfo | undefined {
    return this.activeToolCalls.get(index);
  }

  getToolCallByUseId(toolUseId: string): ToolCallInfo | undefined {
    const index = this.toolUseIdToIndex.get(toolUseId);
    if (index === undefined) return undefined;
    return this.activeToolCalls.get(index);
  }

  appendToolInput(index: number, delta: string): void {
    const current = this.activeToolInputs.get(index) || '';
    this.activeToolInputs.set(index, current + delta);
  }

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

  getToolName(toolUseId: string): string {
    const toolInfo = this.getToolCallByUseId(toolUseId);
    return toolInfo?.tool_name || '';
  }
}
```

## Usage in Handlers

```typescript
// Before
ctx.hasReceivedStreamContent = true;
ctx.activeToolCalls.set(index, toolInfo);
ctx.toolUseIdToIndex.set(toolInfo.tool_use_id, index);

// After
ctx.markStreamContentReceived();
ctx.registerToolCall(index, toolInfo);
```

## Tasks

- [ ] Create `src/context/` directory
- [ ] Create `ExecutionContext.ts` class
- [ ] Implement all methods with proper encapsulation
- [ ] Update `executePrompt` to use class
- [ ] Update all handlers to use context methods
- [ ] Add unit tests for context methods
- [ ] Remove direct Map access throughout code

## Acceptance Criteria

- [ ] All Map operations encapsulated in methods
- [ ] Private properties not directly accessible
- [ ] Methods have clear, descriptive names
- [ ] Unit tests cover all context operations

## References

- `ARCHITECTURE_REVIEW.md` section "2. Extract ExecutionContext as Class"
- Current file: `claude-proxy-sdk/src/claude.ts`

## Labels

```
priority: P3
type: refactoring
component: proxy
```
