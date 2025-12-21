# CLAUDE.md - Claude Proxy SDK

This file provides guidance to Claude Code when working with the Claude Proxy SDK component.

## Project Overview

Claude Proxy SDK is a TypeScript/Node.js service that bridges the Go backend with the Claude Agent SDK. It exposes a WebSocket server that handles streaming Claude API requests on behalf of containerized deployments.

**Claude Agent SDK Documentation**: https://platform.claude.com/docs/en/agent-sdk/overview

## Build & Development Commands

```bash
# Development (watch mode with hot reload)
npm run dev

# Build TypeScript to JavaScript
npm run build

# Run compiled output
npm start

# Type checking without emitting
npm run typecheck
```

## Architecture

### Stack
- **Runtime**: Node.js 24+ with ES modules
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

**Production Dependencies** (required for all builds including production):
- `@anthropic-ai/claude-agent-sdk` - Claude API integration
- `@fastify/cors` - CORS middleware
- `@fastify/websocket` - WebSocket support
- `@types/node` - Node.js type definitions (required during TypeScript compilation)
- `@types/ws` - WebSocket type definitions (required during TypeScript compilation)
- `fastify` - HTTP/WebSocket server framework
- `pino-pretty` - Log formatting
- `typescript` - TypeScript compiler
- `ws` - WebSocket implementation

**Development Dependencies**:
- `tsx` - TypeScript runner for development

Note: `@types/node` and `@types/ws` are in dependencies rather than devDependencies because they are required during the TypeScript compilation step (`npm run build`), which occurs during production update procedures.

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
- Generates source maps and declaration files
- Type checking includes Node.js types explicitly

### Production Builds
When the system updates the proxy (via the update endpoint), it performs:
1. `npm install` - installs all dependencies (including type definitions)
2. `npm run build` - compiles TypeScript to JavaScript
3. Service restart - loads the compiled JavaScript

The type definitions must be available during the build step, so they are production dependencies.

## Deployment

The proxy typically runs alongside the backend in a separate container or on the host machine:

```bash
docker run -d -p 9090:9090 \
  -e ANTHROPIC_API_KEY=sk-ant-... \
  -e API_KEY=your-proxy-key \
  claude-proxy-sdk
```

Backend connects via: `CLAUDE_PROXY_URL=ws://proxy-host:9090`
