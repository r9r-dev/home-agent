# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Home Agent is a web chat interface that wraps Claude Code CLI. It consists of a Go backend (Fiber framework) and Svelte/TypeScript frontend communicating via WebSocket for real-time streaming responses.

The backend connects to a Claude Proxy SDK service via WebSocket. The proxy runs on the host machine and uses the Claude Agent SDK to execute Claude commands on behalf of the containerized backend. This architecture allows the container to use Claude without including the CLI in the image.

**Claude Agent SDK Documentation**: https://platform.claude.com/docs/en/agent-sdk/overview

## Build & Development Commands

```bash
# Development (starts both backend and frontend)
./start-dev.sh

# Or manually:
cd backend && go run main.go          # Backend on :8080
cd frontend && npm run dev            # Frontend on :5173 (proxies /ws to backend)

# Build frontend (outputs to backend/public/)
cd frontend && npm run build

# Build backend
cd backend && go build -o home-agent .

# Type check frontend
cd frontend && npm run check

# Run backend tests
cd backend && go test ./...

# Build container image (use 'container' not 'docker' on this system)
container build -t home-agent .

# Run Claude Proxy SDK locally for development
cd claude-proxy-sdk && npm run dev

# Test backend against local proxy
cd backend && CLAUDE_PROXY_URL=ws://localhost:9090 go run .

# Create release (triggers GitHub Actions)
git tag v0.x.x && git push origin v0.x.x
```

## Architecture

### Three Main Components

1. **Backend** (`backend/`) - Go + Fiber HTTP/WebSocket server
2. **Frontend** (`frontend/`) - Svelte 5 SPA with TypeScript, Tailwind CSS v4, shadcn-svelte
3. **Claude Proxy SDK** (`claude-proxy-sdk/`) - TypeScript/Node.js service using Claude Agent SDK to execute Claude commands on behalf of containerized clients

### Frontend Stack (v0.9.0+)
- **Svelte 5** with runes syntax (`$props`, `$state`, `$derived`, `$effect`)
- **Tailwind CSS v4** via `@tailwindcss/vite` plugin
- **shadcn-svelte** UI components (bits-ui primitives)
- **MynaUI icons** via `@iconify/svelte` (e.g., `<Icon icon="mynaui:edit-one" />`)

Layout & Styling Notes (v0.9.1+):
- ChatWindow uses flexbox with `min-h-0` on flex children to enable proper vertical scrolling
- MessageList ScrollArea requires `min-h-0` class for constrained height behavior
- Always-visible scrollbar styled with theme variables in `app.css` for consistent look
- Markdown content uses `white-space: pre-wrap` for proper text spacing preservation
- ScrollArea uses bits-ui data attributes `[data-scroll-area-scrollbar]` and `[data-scroll-area-thumb]` for CSS styling in `app.css`
- Always-visible scrollbar configured via global CSS, not component classes (to survive shadcn updates)

UI Enhancements (v0.10.0+):
- Collapsible sidebar with toggle button (state persisted in localStorage via `sidebarStore`)
- Sidebar split into two sections: actions ("Nouveau chat", "Rechercher") and chat history ("Vos chats")
- Collapsed sidebar width: 64px, expanded: 260px with smooth transition
- Connection badge: black background with white text, green dot when connected
- Conversation titles generated in French via modified backend prompt
- Cal Sans font for conversation titles (local woff2 in `/public/fonts/`)
- v0.10.1: Lighter Cal Sans font weight (.font-cal class), improved scrollbar visibility, Dialog component for delete confirmation

File Upload Feature (v0.11.0, updated v0.16.0):
- Upload images (PNG, JPG, GIF, WebP) and documents (PDF, TXT, MD, JSON, code files)
- Drag & drop or paperclip button in InputBox
- Container mode: files stored in `/workspace/uploads/{uuid}.ext`
  - `/workspace` mounted from host, mapped to `WORKSPACE_PATH` for Claude CLI
- Local dev mode: `./data/uploads/{uuid}.ext`
- GUID-based filenames prevent collisions (no session subdirectory needed)
- Backend reads file content and includes it in Claude prompt:
  - Text files: content embedded directly (max 100KB)
  - Images: path mapped for Claude's Read tool
- Preview attachments before sending, display in message history
- Key files: `handlers/upload.go`, `InputBox.svelte`, `MessageList.svelte`, `services/api.ts`

Settings Feature (v0.12.0):
- Configuration menu accessible via gear icon in sidebar (bottom section)
- Centered modal dialog with two tabs: "Personnalisation" and "Apercu du prompt"
- Custom instructions (max 2000 chars) appended to system prompt
- Preview shows base system prompt + custom instructions
- Settings persisted in SQLite `settings` table (key-value)
- Key files: `SettingsDialog.svelte`, `settingsStore.ts`, `models/database.go`, `services/claude_executor.go`
- Endpoints: `GET /api/settings`, `PUT /api/settings/:key`, `GET /api/system-prompt`

Bug Fixes (v0.12.1):
- Fixed: Conversation titles now generated in French (reinforced prompt)
- Fixed: Image uploads now accessible to Claude via WORKSPACE_PATH mapping
- Fixed: Attachments parsed from `<!-- attachments:... -->` comments in message history
- Fixed: Horizontal separator between consecutive assistant messages
- Fixed: Settings button moved to bottom of sidebar with Separator
- Added: WORKSPACE_PATH environment variable for Docker deployment path mapping

Bug Fixes (v0.13.2):
- Fixed: Response paragraph separation when Claude responds in multiple parts

Session Management Refactor (v0.16.0):
- **Breaking change**: Session IDs now come from Claude SDK, not generated by frontend
- New conversation flow: SDK generates session_id -> backend creates session -> frontend stores ID
- Resume flow: SDK returns new session_id -> backend updates all references (session + messages)
- File uploads no longer use session subdirectories: `/workspace/uploads/{uuid}.ext`
- Frontend starts with `null` session ID, receives it from backend after first message

Thinking Mode Feature (v0.15.0, Issue #6):
- **Extended Thinking**: Display Claude's internal reasoning process
  - Toggle "Mode Thinking" in Claude menu to enable
  - When enabled, passes `--thinking` flag to Claude CLI
  - Thinking blocks displayed in collapsible UI with amber styling
  - Thinking content streams in real-time, stays visible after response
  - Persisted to database with role "thinking"
- Key files:
  - `frontend/src/components/ThinkingBlock.svelte` - Collapsible thinking display
  - `frontend/src/stores/chatStore.ts` - `thinkingEnabled`, `currentThinking` state
  - `backend/handlers/chat.go` - Handles `thinking` response type
  - `claude-proxy-sdk/src/claude.ts` - Handles thinking events from Claude Agent SDK
- WebSocket message type: `{"type": "thinking", "content": "..."}`

Tool Calls Display Feature (Issue #4):
- **Tool Call Visualization**: Display Claude's tool usage inline in message flow
  - Shows tool name, status (running/success/error), and execution time
  - Collapsible blocks with lazy-loaded input/output content
  - Color-coded: blue (running), green (success), red (error)
  - Persisted to database for history viewing
  - Automatic loading when opening existing conversations
- Key files:
  - `frontend/src/components/ToolCallBlock.svelte` - Collapsible tool call display with lazy loading
  - `frontend/src/stores/chatStore.ts` - `activeToolCalls` Map, tool call actions
  - `frontend/src/components/MessageList.svelte` - Renders tool calls inline
  - `frontend/src/components/ChatWindow.svelte` - Handles tool WebSocket events
  - `frontend/src/services/api.ts` - `fetchToolCalls()`, `fetchToolCallDetail()`
  - `backend/handlers/chat.go` - Handles tool events, persistence
  - `backend/models/database.go` - `tool_calls` table and CRUD
  - `claude-proxy-sdk/src/claude.ts` - Captures tool events from Claude Agent SDK
  - `claude-proxy-sdk/src/types.ts` - `ToolCallInfo` interface
- Tool icons (MynaUI):
  - Bash: `mynaui:terminal`
  - Read/Write/Edit: `mynaui:file`
  - Glob/Grep: `mynaui:search`
  - Task/Agent: `mynaui:layers`
  - WebFetch/WebSearch: `mynaui:globe`
- WebSocket message types:
  - `{"type": "tool_start", "tool": {"tool_use_id": "...", "tool_name": "...", "input": {...}}}`
  - `{"type": "tool_progress", "tool": {"tool_use_id": "..."}, "elapsedTimeSeconds": 2.5}`
  - `{"type": "tool_result", "tool": {"tool_use_id": "..."}, "toolOutput": "...", "isError": false}`
  - `{"type": "tool_error", "tool": {"tool_use_id": "..."}, "toolOutput": "...", "isError": true}`
- Endpoints:
  - `GET /api/sessions/:id/tool-calls` - List tool calls for session (metadata only)
  - `GET /api/tool-calls/:tool_use_id` - Get full tool call detail (lazy loading)

Navigation & Memory Feature (v0.13.0):
- **Menubar Component**: New navigation via Menubar (bits-ui) next to the halfred logo
  - "Menu" dropdown with: Model selection (Haiku/Sonnet/Opus), Thinking mode, Memory, Settings
  - Model selector moved from header to Menubar submenu
  - Connection badge remains on the right side of the header
  - Settings accessible from both Menubar and Sidebar
- **Memory Feature** (Issue #3): Persistent memory injected into every conversation
  - CRUD interface for memory entries (title + content)
  - Toggle individual entries on/off
  - Import/export as JSON
  - Memory preview showing formatted context
  - Memory injected in `<user_memory>` tags before custom instructions
- Key files:
  - `frontend/src/lib/components/ui/menubar/` - Menubar component (new)
  - `frontend/src/components/MemoryDialog.svelte` - Memory management UI
  - `frontend/src/stores/memoryStore.ts` - Memory state management
  - `backend/handlers/memory.go` - Memory REST API
  - `backend/models/database.go` - Memory table and CRUD
  - `backend/handlers/chat.go` - Memory injection into prompt
- Endpoints:
  - `GET /api/memory` - List all entries
  - `POST /api/memory` - Create entry
  - `PUT /api/memory/:id` - Update entry
  - `DELETE /api/memory/:id` - Delete entry
  - `GET /api/memory/export` - Export JSON
  - `POST /api/memory/import` - Import JSON

Real-time Logs Feature (v0.17.0):
- **Log Indicator**: Visual indicator in header showing application health
  - Green: No issues (info level only)
  - Orange: Warnings detected
  - Red: Errors detected
  - Badge shows unread log count
- **Log Panel**: Click indicator to view logs in dialog
  - Scrollable list with colored entries by level
  - Timestamps in French format
  - "Effacer les indicateurs" button to reset status
- **In-memory buffer**: Last 100 log entries (not persisted)
- **Real-time streaming**: WebSocket connection for instant updates
- **Captured events**:
  - Database errors (message save failures, session creation)
  - Claude proxy errors
  - Session management warnings
- Key files:
  - `backend/services/logservice.go` - In-memory log buffer with subscribers
  - `backend/handlers/logs.go` - REST and WebSocket endpoints
  - `frontend/src/stores/logStore.ts` - Log state management
  - `frontend/src/components/LogIndicator.svelte` - Header indicator button
  - `frontend/src/components/LogPanel.svelte` - Log viewer dialog
- Endpoints:
  - `GET /api/logs` - Get all logs and current status
  - `GET /api/logs/status` - Get status only (for polling)
  - `POST /api/logs/clear` - Clear warning/error indicators
  - `GET /ws/logs` - WebSocket for real-time log streaming

System Update Feature (v0.18.0):
- **Update Menu**: "Parametres" renamed to "Systeme" with submenus
  - Green pulsing dot when update is available
  - "Parametres" opens settings dialog
  - "Mises a jour" opens update dialog with "Nouveau" badge
- **Update Dialog**: Modal for managing system updates
  - Shows current and available versions for backend and proxy
  - "Verifier" button to check GitHub Releases API
  - "Lancer la mise a jour" button to start update
  - Dual log panels (Backend Docker / Proxy SDK) with terminal styling
  - Real-time log streaming via WebSocket
- **Update Flow**:
  1. Check versions via GitHub Releases API
  2. Update backend: `docker compose pull && up -d`
  3. Update proxy: Run `install.sh` script (service restart)
- Key files:
  - `claude-proxy-sdk/src/update.ts` - Update logic (check, backend, proxy)
  - `backend/handlers/update.go` - Relay handler to proxy SDK
  - `frontend/src/stores/updateStore.ts` - Update state with WebSocket
  - `frontend/src/components/UpdateDialog.svelte` - Update UI
- Endpoints:
  - `GET /api/update/check` - Check for updates
  - `POST /api/update/backend` - Start backend update
  - `POST /api/update/proxy` - Start proxy update
  - `GET /ws/update` - WebSocket for update logs

**Custom Component Modifications (re-apply after shadcn-svelte updates):**
- `scroll-area.svelte`: Add `type = "always"` prop (default) for always-visible scrollbar
- `scroll-area-scrollbar.svelte`: Custom classes for visible scrollbar:
  - Scrollbar: `bg-muted/50`, `w-3` (vertical), `h-3` (horizontal), `p-0.5`
  - Thumb: `bg-muted-foreground/40 hover:bg-muted-foreground/60`

Key directories:
- `src/components/` - App components (ChatWindow, Sidebar, MessageList, etc.)
- `src/lib/components/ui/` - shadcn-svelte components (button, select, badge, etc.)
- `src/stores/` - Svelte stores for state management
- `src/services/` - API and WebSocket clients

### Backend Key Files
- `main.go` - HTTP server, routes, middleware, initializes ProxyClaudeExecutor and LogService
- `handlers/websocket.go` - WebSocket upgrade and message routing
- `handlers/chat.go` - Message processing, coordinates Claude service and session management, memory injection, error logging
- `handlers/upload.go` - File upload endpoint, serves uploaded files, validates MIME types
- `handlers/memory.go` - Memory CRUD API endpoints
- `handlers/logs.go` - Real-time logs REST and WebSocket endpoints
- `services/claude_executor.go` - Interface definition, shared types (ClaudeResponse, MemoryEntry), system prompt
- `services/proxy_claude_executor.go` - Proxy executor (remote execution via WebSocket to Claude Proxy SDK)
- `services/session.go` - Session CRUD, manages session IDs from Claude Agent SDK
- `services/logservice.go` - In-memory log buffer with real-time subscriber notifications
- `models/database.go` - SQLite schema with migrations, memory table and CRUD

### ClaudeExecutor Interface

The `ClaudeExecutor` interface abstracts Claude CLI execution via the proxy:

```go
type ClaudeExecutor interface {
    ExecuteClaude(ctx, prompt, sessionID, isNewSession, model, customInstructions, thinking) (<-chan ClaudeResponse, error)
    GenerateTitleSummary(userMessage, assistantResponse) (string, error)
    TestConnection() error
}
```

Implementation: `ProxyClaudeExecutor` - Connects to Claude Proxy SDK via WebSocket

ClaudeResponse types: `chunk`, `thinking`, `done`, `error`, `session_id`, `tool_start`, `tool_progress`, `tool_result`, `tool_error`

### Key Data Flow
1. User sends message via WebSocket (`type: "message"`, optionally with `thinking: true`)
2. Backend creates/resumes session
3. Backend retrieves enabled memory entries and custom instructions from database
4. Backend combines memory + custom instructions into system prompt context
5. Backend calls ClaudeExecutor with combined context and thinking flag
6. Executor streams responses back as `ClaudeResponse` events (including tool events)
7. Backend forwards as `type: "chunk"`, `type: "thinking"`, or tool event messages to frontend
8. Backend saves messages and tool calls to SQLite, generates summary title using Claude (haiku)
9. Frontend accumulates chunks in store, displays thinking and tool calls in collapsible blocks, updates UI reactively

### Session Management (v0.16.0+)
- Frontend starts with `null` session_id (no longer generates UUIDs)
- Session IDs are provided by the Claude Agent SDK
- **New conversation flow:**
  1. Frontend sends message with no session_id
  2. SDK executes and returns a session_id
  3. Backend creates session in DB with SDK's session_id
  4. Backend returns session_id to frontend
  5. Frontend stores session_id for subsequent messages
- **Resume flow:**
  1. Frontend sends message with existing session_id
  2. SDK resumes with `resume: session_id` and returns a NEW session_id
  3. Backend updates session_id and all related messages in DB
  4. Backend returns new session_id to frontend
- Titles auto-generated via Claude haiku after first response

## WebSocket Protocol

**Client -> Server:**
```json
{"type": "message", "content": "...", "sessionId": "uuid or null", "model": "haiku", "attachments": [], "thinking": false}
```
Note: `sessionId` is null for new conversations, then set after receiving `session_id` from backend. `thinking` enables extended thinking mode.

Attachments format:
```json
{"id": "uuid", "filename": "file.png", "path": "/api/uploads/...", "type": "image|file", "mime_type": "..."}
```

**Server -> Client:**
```json
{"type": "chunk", "content": "..."}      // Streaming response
{"type": "thinking", "content": "..."}   // Thinking content (when thinking mode enabled)
{"type": "done", "sessionId": "..."}     // Response complete
{"type": "session_id", "sessionId": "..."}  // New session created
{"type": "error", "error": "..."}        // Error occurred
// Tool call events
{"type": "tool_start", "tool": {"tool_use_id": "...", "tool_name": "...", "input": {...}}}
{"type": "tool_progress", "tool": {"tool_use_id": "..."}, "elapsedTimeSeconds": 2.5}
{"type": "tool_result", "tool": {"tool_use_id": "..."}, "toolOutput": "...", "isError": false}
{"type": "tool_error", "tool": {"tool_use_id": "..."}, "toolOutput": "...", "isError": true}
```

## REST API Endpoints

- `POST /api/upload` - Upload file (multipart form, returns `UploadedFile`)
- `GET /api/uploads/:filename` - Serve uploaded file
- `DELETE /api/uploads/:id` - Delete uploaded file
- `GET /api/sessions` - List all sessions
- `GET /api/sessions/:id` - Get session details
- `GET /api/sessions/:id/messages` - Get session messages
- `DELETE /api/sessions/:id` - Delete session
- `PATCH /api/sessions/:id/model` - Update session model
- `GET /api/settings` - Get all settings (key-value map)
- `PUT /api/settings/:key` - Update a setting (body: `{"value": "..."}`)
- `GET /api/system-prompt` - Get base system prompt for preview
- `GET /api/memory` - List all memory entries
- `POST /api/memory` - Create memory entry (body: `{"title": "...", "content": "..."}`)
- `GET /api/memory/:id` - Get single memory entry
- `PUT /api/memory/:id` - Update memory entry (body: `{"title": "...", "content": "...", "enabled": bool}`)
- `DELETE /api/memory/:id` - Delete memory entry
- `GET /api/memory/export` - Export all memory entries as JSON
- `POST /api/memory/import` - Import memory entries (body: `{"entries": [...]}`)
- `GET /api/sessions/:id/tool-calls` - List tool calls for session (metadata only)
- `GET /api/tool-calls/:tool_use_id` - Get full tool call detail (for lazy loading)
- `GET /api/logs` - Get all logs and current status
- `GET /api/logs/status` - Get log status only (info/warning/error)
- `POST /api/logs/clear` - Clear warning/error indicators
- `GET /ws/logs` - WebSocket for real-time log streaming
- `GET /api/update/check` - Check for available updates (backend and proxy versions)
- `POST /api/update/backend` - Start backend (Docker) update
- `POST /api/update/proxy` - Start proxy SDK update
- `GET /ws/update` - WebSocket for real-time update log streaming

## Environment Variables

### Backend (Home Agent)
```bash
PORT=8080                       # Backend port
DATABASE_PATH=./data/homeagent.db
PUBLIC_DIR=./public             # Built frontend directory

# Claude Proxy connection (required)
CLAUDE_PROXY_URL=http://192.168.1.100:9090  # Proxy service URL
CLAUDE_PROXY_KEY=your-api-key               # Proxy authentication (optional)

# Workspace path (required for containerized deployment with file uploads)
WORKSPACE_PATH=/home/user/workspace  # Host path where /workspace is mounted
                                     # Container stores in /workspace/uploads
                                     # Claude CLI accesses via WORKSPACE_PATH/uploads
                                     # If not set, uses ./data/uploads (local dev)
```

### Claude Proxy SDK Service
```bash
PORT=9090                       # Port to listen on (default: 9090)
HOST=0.0.0.0                    # Host to bind to (default: 0.0.0.0)
API_KEY=...                     # API key for proxy authentication (optional)
ANTHROPIC_API_KEY=...           # Anthropic API key (optional if using OAuth)
```

### Authentication Options

The Claude Agent SDK supports two authentication methods:

1. **API Key** (recommended for production):
   ```bash
   export ANTHROPIC_API_KEY=sk-ant-api03-...
   ```

2. **OAuth with Claude Pro/Max subscription**:
   ```bash
   claude setup-token  # Creates a 1-year token
   ```
   Token stored in `~/.claude/.credentials.json`

**Priority order**: `ANTHROPIC_API_KEY` > OAuth token > Interactive login

If OAuth token expires, run `claude setup-token` or `claude /login` again.

## Docker Deployment

The Docker image does NOT include Claude CLI. It requires connection to a Claude Proxy SDK service running on the host:

1. Run Claude Proxy SDK on host:
   ```bash
   cd claude-proxy-sdk && npm install && npm start
   ```
2. Run container with proxy URL and workspace mapping:
   ```bash
   container run -d -p 8080:8080 \
     -v /home/user/workspace:/workspace \
     -v /home/user/data:/app/data \
     -e WORKSPACE_PATH=/home/user/workspace \
     -e CLAUDE_PROXY_URL=ws://HOST_IP:9090 \
     -e CLAUDE_PROXY_KEY=your-key \
     home-agent
   ```

File upload path mapping (v0.16.0):
- Container stores files in `/workspace/uploads/{uuid}.ext`
- Volume mount: host `WORKSPACE_PATH` -> container `/workspace`
- Backend maps paths for Claude: `/workspace/...` -> `WORKSPACE_PATH/...`
- Host's Claude CLI accesses files at `WORKSPACE_PATH/uploads/...`
- GUID-based filenames prevent collisions (no session subdirectory needed)

See `claude-proxy-sdk/README.md` for detailed proxy setup.
