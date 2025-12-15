# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Home Agent is a web chat interface that wraps Claude Code CLI. It consists of a Go backend (Fiber framework) and Svelte/TypeScript frontend communicating via WebSocket for real-time streaming responses.

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

# Docker build
docker build -t home-agent .

# Create release (triggers GitHub Actions)
git tag v0.x.x && git push origin v0.x.x
```

## Architecture

### Backend (Go + Fiber)
- `main.go` - HTTP server, routes, middleware, static file serving
- `handlers/websocket.go` - WebSocket upgrade and message routing
- `handlers/chat.go` - Message processing, coordinates Claude service and session management
- `services/claude.go` - Executes Claude Code CLI with `--resume` for session continuity, streams JSON responses
- `services/session.go` - Session CRUD, maps internal session IDs to Claude CLI session IDs
- `models/database.go` - SQLite schema with migrations, sessions and messages tables

### Frontend (Svelte + TypeScript)
- `components/ChatWindow.svelte` - Main layout, integrates Sidebar and chat area
- `components/Sidebar.svelte` - Conversation history list
- `components/MessageList.svelte` - Renders messages with markdown
- `components/InputBox.svelte` - User input with submit handling
- `stores/chatStore.ts` - Reactive state for messages, connection status, typing indicator
- `services/websocket.ts` - WebSocket client with auto-reconnect
- `services/api.ts` - REST API calls for sessions

### Key Data Flow
1. User sends message via WebSocket (`type: "message"`)
2. Backend creates/resumes session, calls Claude CLI with `--resume <claude_session_id>`
3. Claude CLI streams JSON events, backend forwards as `type: "chunk"` messages
4. Backend saves messages to SQLite, generates summary title using Claude (haiku)
5. Frontend accumulates chunks in store, updates UI reactively

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

```bash
ANTHROPIC_API_KEY=sk-ant-...   # Required for Claude CLI
PORT=8080                       # Backend port
DATABASE_PATH=./data/homeagent.db
CLAUDE_BIN=claude               # Path to Claude CLI binary
PUBLIC_DIR=./public             # Built frontend directory
```

## Docker

- Multi-stage build: Node.js for frontend, Go for backend, Alpine runtime
- Claude CLI installed via npm in container
- Requires CGO for SQLite (`CGO_ENABLED=1`)
- Published to `ghcr.io/r9r-dev/home-agent`
