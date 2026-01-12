# CLAUDE.md - Claude Proxy SDK

This file provides guidance to Claude Code when working with the Claude Proxy SDK component.

## Project Overview

Claude Proxy SDK is a TypeScript service that bridges the Go backend with the Claude Agent SDK. It exposes a WebSocket server that handles streaming Claude API requests on behalf of containerized deployments.

**Runtime**: Bun (executes TypeScript natively, no compilation needed)

**Claude Agent SDK Documentation**: https://platform.claude.com/docs/en/agent-sdk/overview

## Build & Development Commands

```bash
# Development (watch mode with hot reload)
bun run dev

# Run in production
bun start

# Type checking without emitting
bun run typecheck
```

Note: No build step required - Bun executes TypeScript directly.

## Architecture

### Stack
- **Runtime**: Bun with native TypeScript execution
- **Framework**: Fastify 5 for HTTP/WebSocket server
- **SDK**: @anthropic-ai/claude-agent-sdk for Claude integration
- **Language**: TypeScript 5.7 with strict mode
- **WebSocket**: @fastify/websocket for real-time communication
- **Logging**: pino-pretty for development logging

### Key Files
- `src/index.ts` - Fastify server setup, WebSocket handler routing
- `src/claude.ts` - Claude Agent SDK integration, message streaming
- `src/update.ts` - Update logic coordinating backend and proxy updates
- `src/types.ts` - TypeScript interfaces for protocol messages

### Dependencies

- `@anthropic-ai/claude-agent-sdk` - Claude API integration
- `@fastify/cors` - CORS middleware
- `@fastify/websocket` - WebSocket support
- `@types/node` - Node.js type definitions
- `@types/ws` - WebSocket type definitions
- `fastify` - HTTP/WebSocket server framework
- `pino-pretty` - Log formatting
- `typescript` - For type checking only
- `ws` - WebSocket implementation

## Environment Variables

```bash
PORT=9090                    # WebSocket server port
HOST=0.0.0.0                # Bind address
API_KEY=...                 # Authentication key for backend requests
ANTHROPIC_API_KEY=...       # Claude API authentication (sk-ant-... or OAuth token)
```

### Authentication Priority
1. `ANTHROPIC_API_KEY` environment variable
2. OAuth token from `~/.claude/.credentials.json`
3. Interactive login prompt

## WebSocket Protocol

**Connection**: Backend connects to `ws://proxy-host:9090/ws`

**Message Format**:
```json
{
  "type": "message|resume",
  "sessionId": "uuid or null",
  "prompt": "...",
  "model": "haiku|sonnet|opus",
  "customInstructions": "...",
  "thinking": false
}
```

**Response Stream**:
```json
{"type": "chunk", "content": "..."}
{"type": "thinking", "content": "..."}
{"type": "thinking_end"}
{"type": "session_id", "sessionId": "..."}
{"type": "tool_start", "tool": {...}}
{"type": "tool_progress", "tool": {...}, "elapsedTimeSeconds": 2.5}
{"type": "tool_result", "tool": {...}, "toolOutput": "..."}
{"type": "tool_error", "tool": {...}, "toolOutput": "..."}
{"type": "done"}
{"type": "error", "error": "..."}
```

## Development Notes

### TypeScript Configuration
- Target: ES2022
- Module Resolution: NodeNext (ESM)
- Strict mode enabled
- tsconfig.json kept for type checking only (`bun run typecheck`)

### Production Updates
When the system updates the proxy (via the update endpoint), it performs:
1. `bun install` - installs all dependencies
2. Service restart - Bun loads TypeScript directly

No compilation step required.

## Deployment

The proxy typically runs alongside the backend on the host machine:

```bash
cd claude-proxy-sdk
bun install
bun start
```

Backend connects via: `CLAUDE_PROXY_URL=ws://proxy-host:9090`
