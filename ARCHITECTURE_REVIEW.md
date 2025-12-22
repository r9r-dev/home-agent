# Architecture Review & Refactoring Recommendations

This document provides a comprehensive analysis of the Home Agent codebase with specific recommendations for clean code improvements.

## Executive Summary

The codebase is functional but has accumulated technical debt. Key issues:
- **God objects**: Several files exceed 500+ lines with mixed responsibilities
- **Code duplication**: Types defined multiple times across layers
- **Missing abstractions**: Direct dependencies instead of interfaces
- **Inconsistent patterns**: Different approaches for similar problems
- **No shared type contracts**: Frontend/backend types diverge

## Priority Matrix

| Issue | Impact | Effort | Priority |
|-------|--------|--------|----------|
| Split `database.go` into repositories | High | Medium | **P1** |
| Extract `chat.go` prompt builders | High | Medium | **P1** |
| Create shared type definitions | Medium | Low | **P1** |
| Add interfaces for testability | High | Medium | **P2** |
| Split `ChatWindow.svelte` | Medium | Medium | **P2** |
| Consolidate error handling | Medium | Low | **P2** |
| Extract configuration | Low | Low | **P3** |
| Add OpenAPI spec | Medium | High | **P3** |

---

## Backend Recommendations

### 1. Split `models/database.go` (1441 lines) into Domain Repositories

**Current State**: Single 1400+ line file handling all database operations.

**Recommendation**: Create a `repositories/` package with domain-specific files:

```
backend/
├── repositories/
│   ├── repository.go        # Common interface & base
│   ├── session_repo.go      # Session CRUD (~150 lines)
│   ├── message_repo.go      # Message CRUD (~100 lines)
│   ├── memory_repo.go       # Memory entries (~150 lines)
│   ├── machine_repo.go      # SSH machines (~200 lines)
│   ├── tool_call_repo.go    # Tool calls (~100 lines)
│   ├── settings_repo.go     # Settings (~80 lines)
│   └── search_repo.go       # FTS search (~80 lines)
├── models/
│   ├── session.go           # Session struct
│   ├── message.go           # Message struct
│   ├── memory.go            # MemoryEntry struct
│   ├── machine.go           # Machine struct
│   ├── tool_call.go         # ToolCall struct
│   └── search.go            # SearchResult struct
```

**Example Repository Interface**:

```go
// repositories/repository.go
package repositories

type SessionRepository interface {
    Create(sessionID, model string) (*models.Session, error)
    Get(sessionID string) (*models.Session, error)
    List() ([]*models.Session, error)
    UpdateTitle(sessionID, title string) error
    UpdateModel(sessionID, model string) error
    UpdateActivity(sessionID string) error
    Delete(sessionID string) error
}

type MessageRepository interface {
    Save(sessionID, role, content string) (*models.Message, error)
    GetBySession(sessionID string) ([]*models.Message, error)
    Search(query string, limit, offset int) ([]*models.SearchResult, int, error)
}
```

### 2. Refactor `handlers/chat.go` (696 lines)

**Current State**: Single handler mixing prompt building, attachment handling, SSH context, memory injection, and response processing.

**Recommendation**: Extract into focused components:

```
backend/
├── handlers/
│   └── chat.go              # Slim handler (~150 lines)
├── services/
│   ├── prompt/
│   │   ├── builder.go       # PromptBuilder interface
│   │   ├── attachment.go    # AttachmentProcessor
│   │   ├── memory.go        # MemoryInjector
│   │   └── ssh_context.go   # SSHContextBuilder
│   └── response/
│       └── processor.go     # ResponseProcessor
```

**Example PromptBuilder**:

```go
// services/prompt/builder.go
package prompt

type Builder struct {
    attachmentProcessor AttachmentProcessor
    memoryInjector      MemoryInjector
    sshContextBuilder   SSHContextBuilder
}

func (b *Builder) Build(request MessageRequest) (*BuiltPrompt, error) {
    prompt := &BuiltPrompt{}

    // Each step is now testable independently
    if err := b.attachmentProcessor.Process(request, prompt); err != nil {
        return nil, err
    }

    if err := b.memoryInjector.Inject(prompt); err != nil {
        return nil, err
    }

    if request.MachineID != "" {
        if err := b.sshContextBuilder.AddContext(request.MachineID, prompt); err != nil {
            return nil, err
        }
    }

    return prompt, nil
}
```

### 3. Clean Up `main.go` (388 lines)

**Current Issues**:
- Inline route handlers instead of delegation
- Mixed concerns (config, routing, middleware)
- No dependency injection

**Recommendation**: Extract routing and use wire pattern:

```go
// main.go - reduced to ~80 lines
func main() {
    config := loadConfig()

    // Wire up dependencies
    app, cleanup, err := wire.Initialize(config)
    if err != nil {
        log.Fatal(err)
    }
    defer cleanup()

    // Start server with graceful shutdown
    server.Run(app, config.Port)
}
```

```go
// internal/server/routes.go
func RegisterRoutes(app *fiber.App, h *Handlers) {
    // API routes grouped by domain
    api := app.Group("/api")

    sessions := api.Group("/sessions")
    sessions.Get("/", h.Session.List)
    sessions.Get("/:id", h.Session.Get)
    sessions.Delete("/:id", h.Session.Delete)
    // ...
}
```

### 4. Consolidate Type Definitions

**Current Issues**:
- `Attachment` in `websocket.go`
- `MessageAttachment` in `chat.go`
- `ToolInfo` in `chat.go`
- `ToolCallInfo` in `claude_executor.go`
- `ProxyToolInfo` in `proxy_claude_executor.go`

**Recommendation**: Create `types/` package:

```go
// types/attachment.go
package types

type Attachment struct {
    ID       string `json:"id"`
    Filename string `json:"filename"`
    Path     string `json:"path"`
    Type     string `json:"type"` // "image" or "file"
    MimeType string `json:"mime_type,omitempty"`
}

// types/tool.go
type ToolInfo struct {
    ToolUseID       string                 `json:"tool_use_id"`
    ToolName        string                 `json:"tool_name"`
    Input           map[string]interface{} `json:"input,omitempty"`
    ParentToolUseID string                 `json:"parent_tool_use_id,omitempty"`
}
```

### 5. Add Domain Errors

**Current State**: Raw `fmt.Errorf` throughout.

**Recommendation**: Create typed errors:

```go
// errors/errors.go
package errors

var (
    ErrSessionNotFound = &DomainError{Code: "SESSION_NOT_FOUND", Status: 404}
    ErrMachineNotFound = &DomainError{Code: "MACHINE_NOT_FOUND", Status: 404}
    ErrInvalidInput    = &DomainError{Code: "INVALID_INPUT", Status: 400}
    ErrProxyConnection = &DomainError{Code: "PROXY_CONNECTION_FAILED", Status: 503}
)

type DomainError struct {
    Code    string
    Status  int
    Message string
    Cause   error
}

func (e *DomainError) Error() string {
    if e.Message != "" {
        return e.Message
    }
    return e.Code
}

func (e *DomainError) WithMessage(msg string) *DomainError {
    return &DomainError{Code: e.Code, Status: e.Status, Message: msg, Cause: e.Cause}
}
```

### 6. Extract Configuration with Validation

```go
// config/config.go
package config

type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Proxy    ProxyConfig
    Upload   UploadConfig
}

type ServerConfig struct {
    Port      int    `env:"PORT" default:"8080"`
    PublicDir string `env:"PUBLIC_DIR" default:"./public"`
}

type ProxyConfig struct {
    URL    string `env:"CLAUDE_PROXY_URL" required:"true"`
    APIKey string `env:"CLAUDE_PROXY_KEY"`
}

func Load() (*Config, error) {
    cfg := &Config{}
    if err := envconfig.Process("", cfg); err != nil {
        return nil, err
    }
    return cfg, cfg.Validate()
}

func (c *Config) Validate() error {
    if c.Proxy.URL == "" {
        return errors.New("CLAUDE_PROXY_URL is required")
    }
    // More validation...
    return nil
}
```

### 7. Add Interfaces for Testability

**Current Issue**: Handlers directly depend on `*models.DB`.

**Recommendation**:

```go
// Current (hard to test)
type MemoryHandler struct {
    db *models.DB
}

// Better (testable)
type MemoryHandler struct {
    repo repositories.MemoryRepository
}

// Now you can mock in tests
type mockMemoryRepo struct{}
func (m *mockMemoryRepo) Create(id, title, content string) (*models.MemoryEntry, error) {
    return &models.MemoryEntry{ID: id, Title: title, Content: content}, nil
}
```

---

## Frontend Recommendations

### 1. Split `ChatWindow.svelte` (603 lines)

**Current Responsibilities**:
- WebSocket lifecycle management
- Session management
- Dialog state management
- Message sending
- Search integration
- Model/thinking settings
- Keyboard shortcuts

**Recommendation**: Extract into focused components:

```
frontend/src/
├── components/
│   ├── chat/
│   │   ├── ChatWindow.svelte     # Slim coordinator (~100 lines)
│   │   ├── ChatHeader.svelte     # Menubar, logo, indicators
│   │   ├── ChatContent.svelte    # Message list + input wrapper
│   │   ├── EmptyState.svelte     # Welcome screen
│   │   └── DialogManager.svelte  # Dialog state coordination
│   ├── layout/
│   │   └── AppLayout.svelte      # Overall layout structure
```

**Example Refactored ChatWindow**:

```svelte
<script lang="ts">
  import { useWebSocket } from '../hooks/useWebSocket';
  import { useChatSession } from '../hooks/useChatSession';
  import ChatHeader from './ChatHeader.svelte';
  import ChatContent from './ChatContent.svelte';
  import DialogManager from './DialogManager.svelte';

  const { isConnected, sendMessage } = useWebSocket();
  const { session, loadSession, newConversation } = useChatSession();
</script>

<div class="flex h-screen overflow-hidden">
  <Sidebar {session} onSelect={loadSession} onNew={newConversation} />

  <main class="flex-1 flex flex-col">
    <ChatHeader />
    <ChatContent {isConnected} onSend={sendMessage} />
  </main>

  <DialogManager />
</div>
```

### 2. Create Custom Hooks for Logic Extraction

```typescript
// hooks/useWebSocket.ts
export function useWebSocket() {
  const connected = writable(false);

  onMount(() => {
    websocketService.onOpen(() => connected.set(true));
    websocketService.onClose(() => connected.set(false));
    websocketService.connect();

    return () => websocketService.disconnect();
  });

  return {
    isConnected: { subscribe: connected.subscribe },
    sendMessage: websocketService.sendMessage.bind(websocketService),
  };
}

// hooks/useChatSession.ts
export function useChatSession() {
  async function loadSession(sessionId: string) {
    const [session, messages, toolCalls] = await Promise.all([
      fetchSession(sessionId),
      fetchMessages(sessionId),
      fetchToolCalls(sessionId),
    ]);
    // Process and load into store...
  }

  return { loadSession, newConversation: () => chatStore.reset() };
}
```

### 3. Consolidate Type Definitions

**Current Duplication**:
- `MessageAttachment` in `chatStore.ts`
- `MessageAttachment` in `websocket.ts`
- `Attachment` in api types

**Recommendation**: Create shared types:

```typescript
// types/index.ts
export type AttachmentType = 'image' | 'file';

export interface Attachment {
  id: string;
  filename: string;
  path: string;
  type: AttachmentType;
  mimeType?: string;
}

export type MessageRole = 'user' | 'assistant' | 'thinking';
export type ClaudeModel = 'haiku' | 'sonnet' | 'opus';
export type ToolCallStatus = 'running' | 'success' | 'error';

export interface Message {
  id: string;
  role: MessageRole;
  content: string;
  timestamp: Date;
  attachments?: Attachment[];
}
```

### 4. Extract Constants

**Current Issues**: Magic strings scattered throughout.

```typescript
// constants/models.ts
export const CLAUDE_MODELS = {
  HAIKU: 'haiku',
  SONNET: 'sonnet',
  OPUS: 'opus',
} as const;

export const MODEL_LABELS: Record<ClaudeModel, string> = {
  haiku: 'Haiku',
  sonnet: 'Sonnet',
  opus: 'Opus',
};

// constants/websocket.ts
export const WS_MESSAGE_TYPES = {
  CHUNK: 'chunk',
  THINKING: 'thinking',
  DONE: 'done',
  ERROR: 'error',
  // ...
} as const;
```

### 5. Split `chatStore.ts` (544 lines)

**Recommendation**: Separate concerns:

```typescript
// stores/messages.ts - Message state only
// stores/session.ts - Session state
// stores/toolCalls.ts - Tool call state
// stores/thinking.ts - Thinking state
// stores/chat.ts - Composed store combining all
```

### 6. Add API Response Types

```typescript
// types/api.ts
export interface ApiSession {
  id: number;
  session_id: string;
  claude_session_id: string;
  title: string;
  model: ClaudeModel;
  created_at: string;
  last_activity: string;
}

export interface ApiMessage {
  id: number;
  session_id: string;
  role: MessageRole;
  content: string;
  created_at: string;
}

// Type guard functions
export function isApiSession(obj: unknown): obj is ApiSession {
  return typeof obj === 'object' && obj !== null && 'session_id' in obj;
}
```

---

## Claude Proxy SDK Recommendations

### 1. Refactor `processMessage` (230+ lines switch statement)

**Recommendation**: Use handler map pattern:

```typescript
// handlers/messageHandlers.ts
type MessageHandler = (message: SDKMessage, ctx: ExecutionContext) => ProxyResponse | null;

const messageHandlers: Record<string, MessageHandler> = {
  system: handleSystemMessage,
  assistant: handleAssistantMessage,
  stream_event: handleStreamEvent,
  tool_progress: handleToolProgress,
  user: handleUserMessage,
  result: () => null,
};

export function processMessage(message: SDKMessage, ctx: ExecutionContext): ProxyResponse | null {
  const handler = messageHandlers[message.type];
  return handler ? handler(message, ctx) : null;
}

// handlers/streamEventHandlers.ts
function handleStreamEvent(message: SDKMessage, ctx: ExecutionContext): ProxyResponse | null {
  const event = message.event;

  const handlers: Record<string, () => ProxyResponse | null> = {
    message_start: () => handleMessageStart(ctx),
    content_block_start: () => handleContentBlockStart(event, ctx),
    content_block_delta: () => handleContentBlockDelta(event, ctx),
    content_block_stop: () => handleContentBlockStop(event, ctx),
  };

  return handlers[event.type]?.() ?? null;
}
```

### 2. Extract ExecutionContext as Class

```typescript
// context/ExecutionContext.ts
export class ExecutionContext {
  private activeToolCalls = new Map<number, ToolCallInfo>();
  private activeToolInputs = new Map<number, string>();
  private toolUseIdToIndex = new Map<string, number>();
  private hasReceivedStreamContent = false;

  resetForNewMessage(): void {
    this.hasReceivedStreamContent = false;
    this.activeToolCalls.clear();
    this.activeToolInputs.clear();
    this.toolUseIdToIndex.clear();
  }

  registerToolCall(index: number, toolInfo: ToolCallInfo): void {
    this.activeToolCalls.set(index, toolInfo);
    this.toolUseIdToIndex.set(toolInfo.tool_use_id, index);
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
}
```

### 3. Externalize System Prompt

```typescript
// config/systemPrompt.ts
export const SYSTEM_PROMPT = `
You are a helpful personal assistant named Halfred...
`;

// Or load from file
import { readFileSync } from 'fs';
export const SYSTEM_PROMPT = readFileSync('./prompts/system.md', 'utf-8');
```

---

## Cross-Cutting Recommendations

### 1. Create Shared API Contract

Generate TypeScript types from Go structs or use shared JSON Schema:

```yaml
# api/openapi.yaml
openapi: 3.0.0
info:
  title: Home Agent API
  version: 1.0.0

paths:
  /api/sessions:
    get:
      summary: List sessions
      responses:
        '200':
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Session'

components:
  schemas:
    Session:
      type: object
      properties:
        id:
          type: integer
        session_id:
          type: string
        title:
          type: string
        model:
          type: string
          enum: [haiku, sonnet, opus]
```

### 2. Standardize Logging

```go
// Use structured logging
type Logger interface {
    Info(msg string, fields ...Field)
    Warn(msg string, fields ...Field)
    Error(msg string, fields ...Field)
}

// Replace scattered log.Printf with structured logging
logger.Info("session created",
    Field("session_id", sessionID),
    Field("model", model),
)
```

### 3. Add Unit Tests

Focus testing on:
- Repository layer (use SQLite in-memory)
- Prompt building logic
- Response processing
- Store mutations

```go
// repositories/session_repo_test.go
func TestSessionRepository_Create(t *testing.T) {
    db := setupTestDB(t)
    repo := NewSessionRepository(db)

    session, err := repo.Create("test-id", "haiku")

    assert.NoError(t, err)
    assert.Equal(t, "test-id", session.SessionID)
    assert.Equal(t, "haiku", session.Model)
}
```

### 4. Language Consistency

Currently mixing French and English. Choose one:

**Option A**: Full French (current UI language)
```go
// All user-facing strings in French
const ErrSessionNotFound = "Session introuvable"
```

**Option B**: English code, French UI
```go
// Code in English, translations for UI
i18n.T("error.session_not_found")
```

### 5. Extract Magic Numbers

```go
// constants/limits.go
const (
    MaxCustomInstructionsLength = 2000
    MaxFileSizeBytes           = 10 * 1024 * 1024 // 10MB
    MaxFileContentBytes        = 100 * 1024       // 100KB for prompt
    LogBufferSize              = 100
    WebSocketBufferSize        = 100
    ClaudeTimeout              = 10 * time.Minute
)
```

---

## Proposed Directory Structure

```
backend/
├── cmd/
│   └── server/
│       └── main.go              # Slim entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration loading
│   ├── server/
│   │   ├── server.go            # HTTP server setup
│   │   └── routes.go            # Route registration
│   ├── handlers/
│   │   ├── chat.go              # Slim chat handler
│   │   ├── session.go           # Session endpoints
│   │   ├── memory.go            # Memory endpoints
│   │   └── ...
│   ├── services/
│   │   ├── claude/
│   │   │   ├── executor.go      # ClaudeExecutor interface
│   │   │   └── proxy.go         # Proxy implementation
│   │   ├── prompt/
│   │   │   ├── builder.go       # Prompt building
│   │   │   └── context.go       # Context injection
│   │   └── session/
│   │       └── manager.go       # Session management
│   ├── repositories/
│   │   ├── session.go
│   │   ├── message.go
│   │   └── ...
│   └── models/
│       ├── session.go
│       ├── message.go
│       └── ...
├── pkg/
│   ├── errors/
│   │   └── errors.go            # Domain errors
│   └── types/
│       └── types.go             # Shared types
└── go.mod

frontend/
├── src/
│   ├── components/
│   │   ├── chat/
│   │   │   ├── ChatWindow.svelte
│   │   │   └── ...
│   │   ├── layout/
│   │   └── ui/                  # shadcn components
│   ├── hooks/
│   │   ├── useWebSocket.ts
│   │   └── useChatSession.ts
│   ├── stores/
│   │   ├── messages.ts
│   │   ├── session.ts
│   │   └── ...
│   ├── services/
│   │   ├── api.ts
│   │   └── websocket.ts
│   ├── types/
│   │   ├── index.ts             # Shared types
│   │   └── api.ts               # API response types
│   └── constants/
│       └── index.ts
└── package.json

claude-proxy-sdk/
├── src/
│   ├── index.ts
│   ├── claude/
│   │   ├── executor.ts          # Claude execution
│   │   └── context.ts           # Execution context
│   ├── handlers/
│   │   ├── message.ts           # Message handlers
│   │   └── stream.ts            # Stream event handlers
│   ├── types/
│   │   └── index.ts
│   └── config/
│       └── systemPrompt.ts
└── package.json
```

---

## Implementation Roadmap

### Phase 1: Foundation (Week 1-2)
1. Create `types/` package in backend
2. Split `models/database.go` into repositories
3. Add repository interfaces
4. Create shared frontend types

### Phase 2: Core Refactoring (Week 3-4)
1. Extract prompt building from `chat.go`
2. Split `ChatWindow.svelte`
3. Create custom hooks
4. Refactor `processMessage` in proxy SDK

### Phase 3: Polish (Week 5-6)
1. Add domain errors
2. Standardize logging
3. Extract constants
4. Add unit tests for new components

### Phase 4: Documentation (Week 7)
1. Create OpenAPI spec
2. Document WebSocket protocol
3. Update CLAUDE.md

---

## Quick Wins (Can Do Now)

1. **Extract constants** - 30 min
2. **Consolidate frontend types** - 1 hour
3. **Add domain error types** - 1 hour
4. **Split models into separate files** - 2 hours
5. **Extract `PromptBuilder` service** - 2 hours

---

## Conclusion

The codebase is well-structured for a project of this size but would benefit from:

1. **Better separation of concerns** - Split large files
2. **Dependency injection** - Use interfaces for testability
3. **Type consolidation** - Single source of truth for types
4. **Consistent patterns** - Same approach across similar problems

These improvements will make the code more maintainable, testable, and easier to extend.
