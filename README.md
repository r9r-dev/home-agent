# Home Agent

A web-based chat interface for interacting with Claude Code. This project provides a modern, responsive frontend built with Svelte and a Go backend that handles WebSocket connections and Claude API integration.

## Architecture

```
home-agent/
├── backend/           # Go backend server
│   ├── main.go       # Entry point
│   ├── handlers/     # HTTP and WebSocket handlers
│   ├── services/     # Business logic (Claude API, session management)
│   └── models/       # Data models and database
└── frontend/         # Svelte + TypeScript frontend
    ├── src/
    │   ├── components/    # UI components
    │   ├── services/      # WebSocket service
    │   └── stores/        # State management
    └── public/       # Built assets (generated)
```

## Features

### Frontend
- Real-time chat interface with WebSocket
- Markdown rendering with syntax highlighting
- Dark theme optimized for development
- Automatic reconnection with exponential backoff
- Responsive design for mobile and desktop
- Typing indicators and connection status
- Stream responses in real-time

### Backend
- WebSocket server for real-time communication
- Claude API integration
- Session management
- Message history storage
- Static file serving

## Prerequisites

- **Go 1.21+** for backend
- **Node.js 18+** and npm for frontend
- **Claude API key** (set in environment)

## Quick Start

### 1. Setup Environment

Create a `.env` file in the project root:

```bash
cp .env.example .env
```

Edit `.env` and add your configuration:

```env
# Claude API
ANTHROPIC_API_KEY=your_api_key_here

# Server
PORT=8080
HOST=localhost

# Database (optional)
DB_PATH=./home-agent.db
```

### 2. Backend Setup

```bash
cd backend
go mod download
go build
./backend
```

The backend will start on `http://localhost:8080`

### 3. Frontend Setup

```bash
cd frontend
npm install
```

**Development mode:**
```bash
npm run dev
```
Frontend dev server runs on `http://localhost:5173`

**Production build:**
```bash
npm run build
```
Builds to `backend/public/` for serving by the Go backend

## Development Workflow

### Frontend Development

1. Start the backend server (for WebSocket connection)
2. Run `npm run dev` in the frontend directory
3. Access the dev server at `http://localhost:5173`
4. Changes hot-reload automatically

### Backend Development

1. Make changes to Go code
2. Rebuild: `go build`
3. Restart the server

### Full Stack Development

1. Start backend: `cd backend && go run main.go`
2. Start frontend dev server: `cd frontend && npm run dev`
3. Access via `http://localhost:5173` (proxies WebSocket to backend)

## Production Deployment

### Build Frontend

```bash
cd frontend
npm run build
```

This builds optimized assets to `backend/public/`

### Build Backend

```bash
cd backend
go build -o home-agent
```

### Run

```bash
cd backend
./home-agent
```

Access the app at `http://localhost:8080`

## WebSocket Protocol

### Client to Server

```json
{
  "type": "message",
  "content": "User message text"
}
```

### Server to Client

**Streaming chunk:**
```json
{
  "type": "chunk",
  "content": "Response text chunk"
}
```

**Response complete:**
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

## Technology Stack

### Frontend
- **Svelte 4** - Reactive UI framework
- **TypeScript** - Type safety
- **Vite** - Fast build tool
- **Marked** - Markdown parsing
- **Highlight.js** - Code syntax highlighting
- **WebSocket API** - Real-time communication

### Backend
- **Go 1.21+** - Server runtime
- **Gorilla WebSocket** - WebSocket library
- **Claude API** - AI integration
- **SQLite** - Session storage (optional)

## Configuration

### Frontend Environment Variables

Create `frontend/.env`:

```env
VITE_WS_URL=ws://localhost:8080/ws
```

### Backend Environment Variables

Create `backend/.env` or set in environment:

```env
ANTHROPIC_API_KEY=sk-ant-...
PORT=8080
HOST=localhost
DB_PATH=./home-agent.db
```

## Project Structure

### Frontend

```
frontend/
├── src/
│   ├── components/
│   │   ├── ChatWindow.svelte    # Main container
│   │   ├── MessageList.svelte   # Message display
│   │   └── InputBox.svelte      # Input field
│   ├── services/
│   │   └── websocket.ts         # WebSocket manager
│   ├── stores/
│   │   └── chatStore.ts         # State management
│   ├── App.svelte               # Root component
│   ├── main.ts                  # Entry point
│   └── app.css                  # Global styles
├── index.html
├── vite.config.ts
├── tsconfig.json
└── package.json
```

### Backend

```
backend/
├── main.go                      # Entry point
├── handlers/
│   ├── websocket.go            # WebSocket handler
│   └── chat.go                 # Chat logic
├── services/
│   ├── claude.go               # Claude API client
│   └── session.go              # Session management
├── models/
│   └── database.go             # Data models
├── public/                     # Static files (from frontend build)
└── go.mod
```

## API Endpoints

### HTTP
- `GET /` - Serve frontend (index.html)
- `GET /assets/*` - Serve static assets

### WebSocket
- `WS /ws` - WebSocket connection for chat

## Troubleshooting

### Frontend can't connect to backend
- Ensure backend is running on port 8080
- Check `VITE_WS_URL` in frontend `.env`
- Verify no firewall blocking WebSocket connections

### Build errors
- Run `npm install` in frontend directory
- Run `go mod download` in backend directory
- Check Node.js and Go versions

### TypeScript errors
- Run `npm run check` to see detailed errors
- Ensure all dependencies are installed

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## License

MIT

## Support

For issues and questions, please open a GitHub issue.
