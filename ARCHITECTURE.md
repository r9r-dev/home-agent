# Architecture Documentation

Technical architecture and design decisions for Home Agent.

## Overview

Home Agent is a full-stack web application that provides a chat interface for interacting with Claude Code. The system uses a client-server architecture with WebSocket communication for real-time messaging.

```
┌─────────────────────────────────────────────────────────┐
│                      Browser                            │
│  ┌───────────────────────────────────────────────────┐ │
│  │         Svelte Frontend (SPA)                     │ │
│  │  ┌─────────────┐  ┌──────────────┐  ┌──────────┐ │ │
│  │  │ Components  │  │   Stores     │  │ Services │ │ │
│  │  │ - ChatWindow│  │ - chatStore  │  │ - WS     │ │ │
│  │  │ - MessageList  │              │  │          │ │ │
│  │  │ - InputBox  │  │              │  │          │ │ │
│  │  └─────────────┘  └──────────────┘  └──────────┘ │ │
│  └───────────────────────────────────────────────────┘ │
└───────────────────┬─────────────────────────────────────┘
                    │ WebSocket (ws://)
                    │ HTTP (static files)
                    ▼
┌─────────────────────────────────────────────────────────┐
│              Go Backend Server                          │
│  ┌───────────────────────────────────────────────────┐ │
│  │            HTTP Server                            │ │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────────┐   │ │
│  │  │ Handlers │  │ Services │  │   Models     │   │ │
│  │  │ - WS     │  │ - Claude │  │ - Session    │   │ │
│  │  │ - Static │  │ - Session│  │ - Message    │   │ │
│  │  └──────────┘  └──────────┘  └──────────────┘   │ │
│  └───────────────────────────────────────────────────┘ │
└───────────────────┬─────────────────────────────────────┘
                    │ HTTPS API
                    ▼
┌─────────────────────────────────────────────────────────┐
│              Claude API (Anthropic)                     │
└─────────────────────────────────────────────────────────┘
```

## Frontend Architecture

### Technology Stack

- **Svelte 4**: Reactive UI framework with minimal runtime
- **TypeScript**: Type-safe development
- **Vite**: Fast build tool with HMR
- **Marked**: Markdown parser
- **Highlight.js**: Code syntax highlighting

### Directory Structure

```
frontend/src/
├── components/         # Svelte components
│   ├── ChatWindow.svelte
│   ├── MessageList.svelte
│   └── InputBox.svelte
├── services/          # Business logic
│   └── websocket.ts
├── stores/            # State management
│   └── chatStore.ts
├── App.svelte         # Root component
├── main.ts            # Entry point
├── app.css            # Global styles
└── vite-env.d.ts      # Type definitions
```

### Component Hierarchy

```
App
└── ChatWindow
    ├── Header (inline)
    ├── MessageList
    │   └── Message (generated per message)
    └── InputBox
```

### State Management

Uses Svelte stores for reactive state:

```typescript
chatStore {
  messages: Message[]
  currentSessionId: string | null
  isConnected: boolean
  isTyping: boolean
  error: string | null
}
```

### Data Flow

1. **User Input** → InputBox component
2. **Send Message** → WebSocket service
3. **Receive Response** → WebSocket handler
4. **Update Store** → chatStore
5. **Reactive Update** → MessageList re-renders

### WebSocket Service

The WebSocket service manages connection lifecycle:

```typescript
WebSocketService {
  - connect()
  - disconnect()
  - sendMessage(content: string)
  - onMessage(handler)
  - onError(handler)
  - onClose(handler)
  - onOpen(handler)
}
```

Features:
- Automatic reconnection with exponential backoff
- Event-based message handling
- Connection state tracking
- Error handling

### Component Communication

- **Props down**: Parent passes data to children
- **Events up**: Children emit events to parents
- **Store across**: Shared state via Svelte stores

## Backend Architecture

### Technology Stack

- **Go 1.21+**: Server runtime
- **Gorilla WebSocket**: WebSocket library
- **Claude API SDK**: AI integration
- **SQLite**: Session storage (optional)

### Directory Structure

```
backend/
├── main.go           # Entry point, HTTP server
├── handlers/         # Request handlers
│   ├── websocket.go  # WebSocket handler
│   └── chat.go       # Chat logic
├── services/         # Business logic
│   ├── claude.go     # Claude API client
│   └── session.go    # Session management
├── models/           # Data models
│   └── database.go   # DB schemas
└── public/           # Static files (built frontend)
```

### Request Flow

1. **HTTP Request** → Router
2. **Handler** → Validates request
3. **Service** → Business logic
4. **External API** → Claude API
5. **Response** → Back to client

### WebSocket Handler

Manages real-time bidirectional communication:

```go
type WSHandler struct {
    - HandleConnection(conn)
    - HandleMessage(msg)
    - SendResponse(resp)
    - BroadcastMessage(msg)
}
```

Features:
- Connection pooling
- Message queuing
- Error handling
- Graceful shutdown

### Claude Service

Interfaces with Claude API:

```go
type ClaudeService struct {
    - SendMessage(message) -> stream
    - StreamResponse(callback)
    - HandleError(error)
}
```

### Session Management

Tracks user sessions and history:

```go
type SessionService struct {
    - CreateSession() -> sessionID
    - GetSession(id) -> session
    - AddMessage(sessionID, message)
    - GetHistory(sessionID) -> messages
}
```

## Communication Protocol

### WebSocket Messages

#### Client → Server

**Send Message:**
```json
{
  "type": "message",
  "content": "User message text"
}
```

#### Server → Client

**Streaming Chunk:**
```json
{
  "type": "chunk",
  "content": "Response text chunk"
}
```

**Response Complete:**
```json
{
  "type": "done",
  "sessionId": "uuid-string"
}
```

**Error:**
```json
{
  "type": "error",
  "message": "Error description"
}
```

**Session Info:**
```json
{
  "type": "session",
  "sessionId": "uuid-string"
}
```

### Message Flow Sequence

```
Client                    Backend                Claude API
  |                          |                         |
  |-- message ------------>  |                         |
  |                          |-- API request -------> |
  |                          |                         |
  |<-- chunk (typing) ----   |<-- stream chunk -----  |
  |<-- chunk ----------       |<-- stream chunk -----  |
  |<-- chunk ----------       |<-- stream chunk -----  |
  |<-- done -----------       |<-- stream end ------   |
  |                          |                         |
```

## Data Models

### Message

```typescript
interface Message {
  id: string;           // UUID
  role: 'user' | 'assistant';
  content: string;      // Message text
  timestamp: Date;      // Creation time
}
```

### Session

```go
type Session struct {
    ID        string
    Messages  []Message
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

## Security Considerations

### Frontend

- **XSS Prevention**: Markdown sanitization
- **CSRF Protection**: Token-based auth (future)
- **Input Validation**: Message length limits
- **Secure WebSocket**: wss:// in production

### Backend

- **API Key Protection**: Environment variables only
- **Rate Limiting**: Request throttling (future)
- **Input Validation**: Message sanitization
- **CORS Policy**: Strict origin checking
- **TLS/SSL**: HTTPS in production

## Performance Optimizations

### Frontend

- **Code Splitting**: Dynamic imports for large components
- **Lazy Loading**: Load highlight.js on demand
- **Virtual Scrolling**: For long message lists (future)
- **Debouncing**: Typing indicators
- **Caching**: Service worker for assets (future)

### Backend

- **Connection Pooling**: Reuse connections
- **Goroutines**: Concurrent message handling
- **Buffered Channels**: Async processing
- **Message Queuing**: Handle bursts
- **Database Indexing**: Fast lookups

## Scalability

### Current Architecture

- Single server instance
- In-memory session storage
- Direct WebSocket connections

### Future Improvements

- **Horizontal Scaling**: Multiple backend instances
- **Load Balancing**: Nginx/HAProxy
- **Redis**: Shared session storage
- **Message Queue**: RabbitMQ/Kafka
- **Database**: PostgreSQL for persistence
- **CDN**: Static asset distribution

## Deployment Architecture

### Development

```
Frontend Dev Server (Vite)
         ↓
    localhost:5173
         ↓ (proxy /ws)
Backend Server (Go)
         ↓
    localhost:8080
```

### Production

```
       Nginx (HTTPS)
            ↓
     Port 443/80
            ↓
    Backend Server (Go)
    Port 8080 (internal)
            ↓
       Claude API
```

## Error Handling

### Frontend

1. **Connection Errors**: Show reconnecting indicator
2. **Send Errors**: Display error message
3. **Parse Errors**: Log and fallback
4. **Network Errors**: Retry with backoff

### Backend

1. **API Errors**: Return error message to client
2. **Connection Errors**: Close gracefully
3. **Validation Errors**: Return 400 status
4. **Server Errors**: Log and return 500

## Monitoring & Logging

### Frontend

- Browser console logs
- Error tracking (Sentry, future)
- Performance monitoring

### Backend

- Structured logging
- Request/response logging
- Error tracking
- Metrics (Prometheus, future)

## Testing Strategy

### Frontend

- **Unit Tests**: Component logic (Jest, future)
- **Integration Tests**: Component interaction
- **E2E Tests**: Full user flows (Playwright, future)
- **Type Checking**: TypeScript strict mode

### Backend

- **Unit Tests**: Service logic
- **Integration Tests**: Handler tests
- **Load Tests**: WebSocket concurrency
- **API Tests**: Mock Claude API

## Future Enhancements

### Short Term

- User authentication
- Session persistence
- Message history
- File uploads

### Long Term

- Multi-user support
- Collaboration features
- Plugin system
- Voice input
- Mobile apps

## Dependencies

### Frontend

```json
{
  "svelte": "^4.2.8",
  "marked": "^11.1.0",
  "highlight.js": "^11.9.0",
  "typescript": "^5.3.3",
  "vite": "^5.0.8"
}
```

### Backend

```
gorilla/websocket
anthropic/claude-sdk-go
```

## Configuration

### Environment Variables

**Frontend:**
- `VITE_WS_URL`: WebSocket URL

**Backend:**
- `ANTHROPIC_API_KEY`: Claude API key
- `PORT`: Server port
- `HOST`: Server host
- `DB_PATH`: Database path

## Build Process

### Development

```bash
# Frontend: Vite dev server with HMR
npm run dev

# Backend: Go with live reload
go run main.go
```

### Production

```bash
# Frontend: Optimized bundle
npm run build
# Output: backend/public/

# Backend: Compiled binary
go build -ldflags="-s -w"
# Output: backend/home-agent
```

## References

- [Svelte Documentation](https://svelte.dev/docs)
- [Go WebSocket Tutorial](https://golang.org/doc/)
- [Claude API Documentation](https://docs.anthropic.com/)
- [WebSocket Protocol](https://tools.ietf.org/html/rfc6455)
