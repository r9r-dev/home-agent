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

### Components

1. **Backend** (`backend/`) - Go + Fiber HTTP/WebSocket server
2. **Frontend** (`frontend/`) - Svelte 5 SPA with TypeScript, Tailwind CSS v4, shadcn-svelte
3. **Claude Proxy SDK** (`claude-proxy-sdk/`) - TypeScript/Node.js service using Claude Agent SDK

### Frontend Structure

**Stack:**
- Svelte 5 with runes syntax (`$props`, `$state`, `$derived`, `$effect`)
- Tailwind CSS v4 via `@tailwindcss/vite` plugin
- shadcn-svelte UI components (bits-ui primitives)
- MynaUI icons via `@iconify/svelte` (e.g., `<Icon icon="mynaui:edit-one" />`)

**Key Directories:**
- `src/components/` - App components (ChatWindow, Sidebar, MessageList, etc.)
- `src/lib/components/ui/` - shadcn-svelte components (button, select, badge, menubar, etc.)
- `src/stores/` - Svelte stores for state management
- `src/services/` - API and WebSocket clients

**Layout Notes:**
- ChatWindow uses flexbox with `min-h-0` on flex children for proper vertical scrolling
- ScrollArea uses bits-ui data attributes for CSS styling in `app.css`
- Always-visible scrollbar configured via global CSS (survives shadcn updates)
- Collapsible sidebar (64px collapsed, 260px expanded) with state persisted in localStorage

**Custom Component Modifications (re-apply after shadcn-svelte updates):**
- `scroll-area.svelte`: Add `type = "always"` prop (default) for always-visible scrollbar
- `scroll-area-scrollbar.svelte`: Custom classes for visible scrollbar

### Backend Structure

**Directory Layout:**
```
backend/
├── handlers/          # HTTP/WebSocket handlers
├── services/          # Business logic services
├── repositories/      # Data access layer (repository pattern)
└── models/            # Data structures and DB initialization
```

**Key Files:**
- `main.go` - HTTP server, routes, middleware, dependency injection
- `handlers/websocket.go` - WebSocket upgrade and message routing
- `handlers/chat.go` - Message processing, session management, memory/SSH injection
- `handlers/upload.go` - File upload endpoint, MIME validation
- `handlers/memory.go` - Memory CRUD API (uses MemoryRepository)
- `handlers/machines.go` - SSH machines CRUD and connection testing
- `handlers/search.go` - Full-text search API (uses SearchRepository)
- `handlers/logs.go` - Real-time logs REST and WebSocket
- `handlers/update.go` - System update relay to proxy
- `services/claude_executor.go` - Interface definition, ClaudeResponse types
- `services/proxy_claude_executor.go` - WebSocket connection to Claude Proxy SDK
- `services/session.go` - Session management (uses SessionRepository, MessageRepository)
- `services/logservice.go` - In-memory log buffer with subscribers
- `services/crypto.go` - AES-256-GCM encryption for SSH credentials
- `services/ssh.go` - SSH connection testing
- `models/database.go` - SQLite initialization and migrations only
- `models/*.go` - Data structures (Session, Message, MemoryEntry, Machine, ToolCall, SearchResult)

**Repositories** (`repositories/`):
- `interfaces.go` - All repository interfaces for dependency injection
- `session_repo.go` - SessionRepository: session CRUD, session ID updates
- `message_repo.go` - MessageRepository: message persistence
- `memory_repo.go` - MemoryRepository: memory entries CRUD
- `machine_repo.go` - MachineRepository: SSH machines with encrypted credentials
- `tool_call_repo.go` - ToolCallRepository: tool call tracking
- `settings_repo.go` - SettingsRepository: key-value settings
- `search_repo.go` - SearchRepository: FTS5 full-text search

**ClaudeExecutor Interface:**
```go
type ClaudeExecutor interface {
    ExecuteClaude(ctx, prompt, sessionID, isNewSession, model, customInstructions, thinking) (<-chan ClaudeResponse, error)
    GenerateTitleSummary(userMessage, assistantResponse) (string, error)
    TestConnection() error
}
```

ClaudeResponse types: `chunk`, `thinking`, `thinking_end`, `done`, `error`, `session_id`, `tool_start`, `tool_progress`, `tool_result`, `tool_error`

### Claude Proxy SDK Structure

**Key Files:**
- `src/index.ts` - Fastify server, WebSocket handling
- `src/claude.ts` - Claude Agent SDK integration, streaming
- `src/update.ts` - Update logic for backend and proxy
- `src/types.ts` - TypeScript interfaces

## Features

### Chat & Conversations
- Real-time streaming responses via WebSocket
- Session management with SDK-generated session IDs
- Conversation history stored in SQLite
- Auto-generated French titles via Claude haiku
- Model selection: Haiku, Sonnet, Opus

### File Uploads
- Supported: images (PNG, JPG, GIF, WebP), documents (PDF, TXT, MD, JSON, code files)
- Drag & drop or paperclip button in InputBox
- GUID-based filenames in `/workspace/uploads/` (container) or `./data/uploads/` (local)
- Text files embedded in prompt (max 100KB), images mapped for Claude's Read tool
- Preview before sending, display in message history

### Settings & Customization
- Custom instructions (max 2000 chars) appended to system prompt
- Collapsible prompt preview in settings dialog
- Settings persisted in SQLite `settings` table

### Memory System
- Persistent memory entries injected into every conversation
- CRUD interface with title + content
- Toggle individual entries on/off
- Import/export as JSON
- Memory injected in `<user_memory>` tags before custom instructions

### Extended Thinking Mode
- Toggle in Claude menu to display Claude's reasoning process
- Thinking blocks in collapsible UI with amber styling
- Multiple thinking blocks displayed separately in chronological order
- Persisted to database with role "thinking"

### Tool Calls Display
- Inline visualization of Claude's tool usage
- Shows tool name, status (running/success/error), execution time
- Collapsible blocks with lazy-loaded input/output
- Color-coded: blue (running), green (success), red (error)
- Tool icons: terminal (Bash), file (Read/Write), search (Glob/Grep), layers (Task), globe (Web)

### SSH Machines
- Remote machine management in Settings → "Connexions SSH"
- Fields: name, description, host, port, username, auth (password or SSH key)
- Test connection with latency display
- Machine selector in InputBox for targeting remote execution
- Credentials encrypted with AES-256-GCM (key derived from database path)

### Real-time Logs
- Log indicator in header: green (OK), orange (warnings), red (errors)
- Log panel with scrollable entries, timestamps, clear button
- In-memory buffer (last 100 entries)
- WebSocket streaming for instant updates

### System Updates
- Check versions via GitHub Releases API
- Update backend: `docker compose pull && up -d`
- Update proxy: git clone + npm build + service restart
- Dual log panels with real-time streaming
- Auto-reconnection during backend restart

## Data Flow

1. User sends message via WebSocket (with optional `thinking`, `machineId`)
2. Backend creates/resumes session
3. Backend retrieves memory entries and custom instructions
4. If `machineId` provided, fetches machine and injects SSH context
5. Backend calls ClaudeExecutor with combined context
6. Executor streams responses as `ClaudeResponse` events
7. Backend forwards chunks, thinking, tool events to frontend
8. Backend saves messages and tool calls to SQLite
9. Backend generates summary title using Claude haiku
10. Frontend accumulates in store, updates UI reactively

### Session Management
- Frontend starts with `null` session_id
- SDK generates session_id on first message
- Resume flow: SDK returns new session_id, backend updates all references
- Titles auto-generated after first response

## WebSocket Protocol

**Client -> Server:**
```json
{
  "type": "message",
  "content": "...",
  "sessionId": "uuid or null",
  "model": "haiku|sonnet|opus",
  "attachments": [{"id": "...", "filename": "...", "path": "...", "type": "image|file", "mime_type": "..."}],
  "thinking": false,
  "machineId": "uuid or null"
}
```

**Server -> Client:**
```json
{"type": "chunk", "content": "..."}
{"type": "thinking", "content": "..."}
{"type": "thinking_end"}
{"type": "done", "sessionId": "..."}
{"type": "session_id", "sessionId": "..."}
{"type": "session_title", "sessionId": "...", "title": "..."}
{"type": "error", "error": "..."}
{"type": "tool_start", "tool": {"tool_use_id": "...", "tool_name": "...", "input": {...}}}
{"type": "tool_progress", "tool": {"tool_use_id": "..."}, "elapsedTimeSeconds": 2.5}
{"type": "tool_result", "tool": {"tool_use_id": "..."}, "toolOutput": "...", "isError": false}
{"type": "tool_error", "tool": {"tool_use_id": "..."}, "toolOutput": "...", "isError": true}
```

## REST API Endpoints

### Sessions
- `GET /api/sessions` - List all sessions
- `GET /api/sessions/:id` - Get session details
- `GET /api/sessions/:id/messages` - Get session messages
- `DELETE /api/sessions/:id` - Delete session
- `PATCH /api/sessions/:id/model` - Update session model

### File Uploads
- `POST /api/upload` - Upload file (multipart form)
- `GET /api/uploads/:filename` - Serve uploaded file
- `DELETE /api/uploads/:id` - Delete uploaded file

### Settings
- `GET /api/settings` - Get all settings
- `PUT /api/settings/:key` - Update setting
- `GET /api/system-prompt` - Get base system prompt

### Memory
- `GET /api/memory` - List all entries
- `POST /api/memory` - Create entry
- `GET /api/memory/:id` - Get entry
- `PUT /api/memory/:id` - Update entry
- `DELETE /api/memory/:id` - Delete entry
- `GET /api/memory/export` - Export as JSON
- `POST /api/memory/import` - Import from JSON

### Tool Calls
- `GET /api/sessions/:id/tool-calls` - List tool calls (metadata only)
- `GET /api/tool-calls/:tool_use_id` - Get full detail (lazy loading)

### SSH Machines
- `GET /api/machines` - List machines (credentials excluded)
- `POST /api/machines` - Create machine
- `GET /api/machines/:id` - Get machine
- `PUT /api/machines/:id` - Update machine
- `DELETE /api/machines/:id` - Delete machine
- `POST /api/machines/:id/test` - Test SSH connection

### Logs
- `GET /api/logs` - Get all logs and status
- `GET /api/logs/status` - Get status only
- `POST /api/logs/clear` - Clear indicators
- `GET /ws/logs` - WebSocket for streaming

### Updates
- `GET /api/update/check` - Check for updates
- `POST /api/update/backend` - Start backend update
- `POST /api/update/proxy` - Start proxy update
- `GET /ws/update` - WebSocket for update logs

## Environment Variables

### Backend
```bash
PORT=8080
DATABASE_PATH=./data/homeagent.db
PUBLIC_DIR=./public

# Claude Proxy (required)
CLAUDE_PROXY_URL=http://192.168.1.100:9090
CLAUDE_PROXY_KEY=your-api-key

# Workspace (for containerized deployment)
WORKSPACE_PATH=/home/user/workspace
```

### Claude Proxy SDK
```bash
PORT=9090
HOST=0.0.0.0
API_KEY=...
ANTHROPIC_API_KEY=...  # Or use OAuth token
```

### Authentication
1. **API Key**: `export ANTHROPIC_API_KEY=sk-ant-api03-...`
2. **OAuth**: `claude setup-token` (creates 1-year token in `~/.claude/.credentials.json`)

Priority: `ANTHROPIC_API_KEY` > OAuth token > Interactive login

## Docker Deployment

The Docker image requires connection to a Claude Proxy SDK service on the host:

```bash
# 1. Run proxy on host
cd claude-proxy-sdk && npm install && npm start

# 2. Run container
container run -d -p 8080:8080 \
  -v /home/user/workspace:/workspace \
  -v /home/user/data:/app/data \
  -e WORKSPACE_PATH=/home/user/workspace \
  -e CLAUDE_PROXY_URL=ws://HOST_IP:9090 \
  -e CLAUDE_PROXY_KEY=your-key \
  home-agent
```

**File Upload Path Mapping:**
- Container stores: `/workspace/uploads/{uuid}.ext`
- Volume: host `WORKSPACE_PATH` -> container `/workspace`
- Claude CLI accesses: `WORKSPACE_PATH/uploads/...`

See `claude-proxy-sdk/README.md` for detailed proxy setup.

## Development Guidelines

### Commit & Release Policy
1. Update CLAUDE.md with changes
2. Commit and push
3. Close related GitHub issue if any

Never publish tag. User will do it instead.