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
- **Lucide icons** (`@lucide/svelte`)

Key directories:
- `src/components/` - App components (ChatWindow, Sidebar, MessageList, etc.)
- `src/lib/components/ui/` - shadcn-svelte components (button, select, badge, etc.)
- `src/stores/` - Svelte stores for state management
- `src/services/` - API and WebSocket clients

### Backend Key Files
- `main.go` - HTTP server, routes, middleware, initializes ClaudeExecutor based on config
- `handlers/websocket.go` - WebSocket upgrade and message routing
- `handlers/chat.go` - Message processing, coordinates Claude service and session management
- `services/claude_executor.go` - Interface definition for Claude execution
- `services/claude.go` - Local executor (direct CLI execution)
- `services/proxy_claude_executor.go` - Proxy executor (remote execution via WebSocket)
- `services/session.go` - Session CRUD, maps internal session IDs to Claude CLI session IDs
- `models/database.go` - SQLite schema with migrations

### ClaudeExecutor Interface

The `ClaudeExecutor` interface abstracts Claude CLI execution:

```go
type ClaudeExecutor interface {
    ExecuteClaude(ctx, prompt, sessionID) (<-chan ClaudeResponse, error)
    GenerateTitleSummary(userMessage, assistantResponse) (string, error)
    TestConnection() error
}
```

Two implementations:
- `LocalClaudeExecutor` - Spawns Claude CLI process directly
- `ProxyClaudeExecutor` - Connects to Claude Proxy via WebSocket

### Key Data Flow
1. User sends message via WebSocket (`type: "message"`)
2. Backend creates/resumes session, calls ClaudeExecutor
3. Executor streams responses back as `ClaudeResponse` events
4. Backend forwards as `type: "chunk"` messages to frontend
5. Backend saves messages to SQLite, generates summary title using Claude (haiku)
6. Frontend accumulates chunks in store, updates UI reactively

### Session Management
- Internal `session_id` (UUID) used for database foreign keys and frontend routing
- `claude_session_id` stored separately for Claude CLI `--resume` flag
- Titles auto-generated via Claude haiku after first response

## WebSocket Protocol

**Client -> Server:**
```json
{"type": "message", "content": "...", "session_id": "optional-uuid"}
```

**Server -> Client:**
```json
{"type": "chunk", "content": "..."}      // Streaming response
{"type": "done", "sessionId": "..."}     // Response complete
{"type": "session_id", "sessionId": "..."}  // New session created
{"type": "error", "error": "..."}        // Error occurred
```

## Environment Variables

### Backend (Home Agent)
```bash
PORT=8080                       # Backend port
DATABASE_PATH=./data/homeagent.db
PUBLIC_DIR=./public             # Built frontend directory

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
2. Run container with proxy URL:
   ```bash
   container run -d -p 8080:8080 \
     -e CLAUDE_PROXY_URL=http://HOST_IP:9090 \
     -e CLAUDE_PROXY_KEY=your-key \
     home-agent
   ```

See `docs/claude-proxy.md` for detailed proxy setup.
