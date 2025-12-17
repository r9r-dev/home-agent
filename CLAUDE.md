# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Home Agent is a web chat interface that wraps Claude Code CLI. It consists of a Go backend (Fiber framework) and Svelte/TypeScript frontend communicating via WebSocket for real-time streaming responses.

The system supports two execution modes:
- **Local mode**: Backend executes Claude CLI directly on the same machine
- **Proxy mode**: Backend connects to a remote Claude Proxy service via WebSocket, allowing the container to use Claude CLI running on the host

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

# Run Claude Proxy locally for development
cd claude-proxy && PROXY_PORT=9090 go run .

# Test backend against local proxy
cd backend && CLAUDE_PROXY_URL=http://localhost:9090 go run .

# Create release (triggers GitHub Actions)
git tag v0.x.x && git push origin v0.x.x
```

## Architecture

### Three Main Components

1. **Backend** (`backend/`) - Go + Fiber HTTP/WebSocket server
2. **Frontend** (`frontend/`) - Svelte 5 SPA with TypeScript, Tailwind CSS v4, shadcn-svelte
3. **Claude Proxy** (`claude-proxy/`) - Standalone service that executes Claude CLI on behalf of containerized clients

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

File Upload Feature (v0.11.0):
- Upload images (PNG, JPG, GIF, WebP) and documents (PDF, TXT, MD, JSON, code files)
- Drag & drop or paperclip button in InputBox
- Files stored in `./data/uploads/{session_id}/`
- Backend reads file content and includes it in Claude prompt:
  - Text files: content embedded directly (max 100KB)
  - Images: absolute path provided for Claude's Read tool
- Preview attachments before sending, display in message history
- Key files: `handlers/upload.go`, `InputBox.svelte`, `MessageList.svelte`, `services/api.ts`

Settings Feature (v0.12.0):
- Configuration menu accessible via gear icon in sidebar (bottom section)
- Centered modal dialog with two tabs: "Personnalisation" and "Apercu du prompt"
- Custom instructions (max 2000 chars) appended to system prompt
- Preview shows base system prompt + custom instructions
- Settings persisted in SQLite `settings` table (key-value)
- Key files: `SettingsDialog.svelte`, `settingsStore.ts`, `models/database.go`, `services/claude.go`
- Endpoints: `GET /api/settings`, `PUT /api/settings/:key`, `GET /api/system-prompt`

Bug Fixes (v0.12.1):
- Fixed: Conversation titles now generated in French (reinforced prompt)
- Fixed: Image uploads now accessible to Claude via WORKSPACE_PATH mapping
- Fixed: Attachments parsed from `<!-- attachments:... -->` comments in message history
- Fixed: Horizontal separator between consecutive assistant messages
- Fixed: Settings button moved to bottom of sidebar with Separator
- Added: WORKSPACE_PATH environment variable for Docker deployment path mapping

Navigation & Memory Feature (v0.13.0):
- **Menubar Component**: New navigation via Menubar (bits-ui) next to the halfred logo
  - "Menu" dropdown with: Model selection (Haiku/Sonnet/Opus), Memory, Settings
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
- `main.go` - HTTP server, routes, middleware, initializes ClaudeExecutor based on config
- `handlers/websocket.go` - WebSocket upgrade and message routing
- `handlers/chat.go` - Message processing, coordinates Claude service and session management, memory injection
- `handlers/upload.go` - File upload endpoint, serves uploaded files, validates MIME types
- `handlers/memory.go` - Memory CRUD API endpoints
- `services/claude_executor.go` - Interface definition for Claude execution
- `services/claude.go` - Local executor (direct CLI execution), memory formatting
- `services/proxy_claude_executor.go` - Proxy executor (remote execution via WebSocket)
- `services/session.go` - Session CRUD, maps internal session IDs to Claude CLI session IDs
- `models/database.go` - SQLite schema with migrations, memory table and CRUD

### ClaudeExecutor Interface

The `ClaudeExecutor` interface abstracts Claude CLI execution:

```go
type ClaudeExecutor interface {
    ExecuteClaude(ctx, prompt, sessionID, model, customInstructions) (<-chan ClaudeResponse, error)
    GenerateTitleSummary(userMessage, assistantResponse) (string, error)
    TestConnection() error
}
```

Two implementations:
- `LocalClaudeExecutor` - Spawns Claude CLI process directly
- `ProxyClaudeExecutor` - Connects to Claude Proxy via WebSocket

### Key Data Flow
1. User sends message via WebSocket (`type: "message"`)
2. Backend creates/resumes session
3. Backend retrieves enabled memory entries and custom instructions from database
4. Backend combines memory + custom instructions into system prompt context
5. Backend calls ClaudeExecutor with combined context
6. Executor streams responses back as `ClaudeResponse` events
7. Backend forwards as `type: "chunk"` messages to frontend
8. Backend saves messages to SQLite, generates summary title using Claude (haiku)
9. Frontend accumulates chunks in store, updates UI reactively

### Session Management
- Internal `session_id` (UUID) used for database foreign keys and frontend routing
- `claude_session_id` stored separately for Claude CLI `--resume` flag
- Titles auto-generated via Claude haiku after first response

## WebSocket Protocol

**Client -> Server:**
```json
{"type": "message", "content": "...", "session_id": "optional-uuid", "attachments": []}
```

Attachments format:
```json
{"id": "uuid", "filename": "file.png", "path": "/api/uploads/...", "type": "image|file", "mime_type": "..."}
```

**Server -> Client:**
```json
{"type": "chunk", "content": "..."}      // Streaming response
{"type": "done", "sessionId": "..."}     // Response complete
{"type": "session_id", "sessionId": "..."}  // New session created
{"type": "error", "error": "..."}        // Error occurred
```

## REST API Endpoints

- `POST /api/upload` - Upload file (multipart form, returns `UploadedFile`)
- `GET /api/uploads/:sessionId/:filename` - Serve uploaded file
- `DELETE /api/uploads/:id?session_id=...` - Delete uploaded file
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
- `POST /api/memory/import` - Import memory entries (body: `{"entries": [...]}`))

## Environment Variables

### Backend (Home Agent)
```bash
PORT=8080                       # Backend port
DATABASE_PATH=./data/homeagent.db
PUBLIC_DIR=./public             # Built frontend directory
UPLOAD_DIR=./data/uploads       # Directory for uploaded files

# Workspace mapping (for containerized deployment)
WORKSPACE_PATH=/home/user/workspace  # Root workspace path that Claude CLI sees
                                     # Backend automatically adds "uploads" subfolder
                                     # Use when backend runs in container but Claude on host

# Local mode (direct CLI execution)
CLAUDE_BIN=claude               # Path to Claude CLI binary
ANTHROPIC_API_KEY=sk-ant-...   # Required for Claude CLI

# Proxy mode (remote execution)
CLAUDE_PROXY_URL=http://192.168.1.100:9090  # Proxy service URL
CLAUDE_PROXY_KEY=your-api-key               # Proxy authentication
```

### Claude Proxy Service
```bash
PROXY_PORT=9090                 # Port to listen on
PROXY_HOST=0.0.0.0              # Host to bind to
PROXY_API_KEY=...               # API key for authentication
CLAUDE_BIN=claude               # Path to Claude CLI
```

## Docker Deployment

The Docker image does NOT include Claude CLI. It requires connection to a Claude Proxy service running on the host:

1. Install Claude Proxy on host: `curl -fsSL .../install.sh | sudo bash`
2. Run container with proxy URL and workspace mapping:
   ```bash
   container run -d -p 8080:8080 \
     -v /home/user/workspace:/workspace \
     -e CLAUDE_PROXY_URL=http://HOST_IP:9090 \
     -e CLAUDE_PROXY_KEY=your-key \
     -e UPLOAD_DIR=/data/uploads \
     -e WORKSPACE_PATH=/home/user/workspace \
     home-agent
   ```

The `WORKSPACE_PATH` variable maps to the root workspace directory on the host:
- Container stores files in `UPLOAD_DIR` (e.g., `/data/uploads/session_id/file.png`)
- Backend extracts relative path and builds Claude path: `WORKSPACE_PATH/uploads/session_id/file.png`
- Host's Claude CLI accesses files via the mounted volume
- The "uploads" subfolder is added automatically, keeping workspace organized for future subfolders

See `docs/claude-proxy.md` for detailed proxy setup.
