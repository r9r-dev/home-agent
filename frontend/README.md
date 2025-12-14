# Home Agent Frontend

Modern web interface for interacting with Claude Code via WebSocket.

## Features

- Real-time chat interface with Claude Code
- WebSocket-based communication with automatic reconnection
- Markdown rendering with syntax highlighting
- Responsive, mobile-friendly design
- Dark theme optimized for long sessions
- Typing indicators and connection status
- Stream responses in real-time

## Tech Stack

- **Svelte 4** - Reactive UI framework
- **TypeScript** - Type-safe development
- **Vite** - Fast build tool
- **Marked** - Markdown parser
- **Highlight.js** - Syntax highlighting for code blocks
- **WebSocket** - Real-time bidirectional communication

## Getting Started

### Prerequisites

- Node.js 18+ and npm
- Backend server running on port 8080 (or configure via environment variable)

### Installation

```bash
npm install
```

### Configuration

Create a `.env` file from the example:

```bash
cp .env.example .env
```

Edit `.env` to configure the WebSocket URL:

```env
VITE_WS_URL=ws://localhost:8080/ws
```

### Development

Start the development server:

```bash
npm run dev
```

The app will be available at `http://localhost:5173` (or another port if 5000 is in use).

### Build for Production

```bash
npm run build
```

The built files will be in the `dist/` directory.

### Preview Production Build

```bash
npm run preview
```

## Project Structure

```
frontend/
├── src/
│   ├── components/          # Svelte components
│   │   ├── ChatWindow.svelte    # Main chat container
│   │   ├── MessageList.svelte   # Message display with markdown
│   │   └── InputBox.svelte      # Message input with auto-resize
│   ├── services/            # Business logic
│   │   └── websocket.ts         # WebSocket connection manager
│   ├── stores/              # Svelte stores
│   │   └── chatStore.ts         # Chat state management
│   ├── App.svelte           # Root component
│   ├── main.ts              # Entry point
│   └── app.css              # Global styles
├── index.html               # HTML template
├── vite.config.ts           # Vite configuration
└── package.json             # Dependencies
```

## Key Components

### WebSocket Service

The WebSocket service (`services/websocket.ts`) handles:
- Connection management with automatic reconnection
- Exponential backoff for reconnection attempts
- Message sending and receiving
- Event handlers for connection lifecycle

### Chat Store

The Svelte store (`stores/chatStore.ts`) manages:
- Message history
- Connection status
- Typing indicators
- Session management
- Error states

### Components

1. **ChatWindow**: Main container that coordinates the chat interface
2. **MessageList**: Displays messages with markdown rendering and syntax highlighting
3. **InputBox**: Message input with auto-resize and keyboard shortcuts

## WebSocket Message Format

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

## Styling

The app uses a dark theme optimized for development:
- CSS custom properties for easy theming
- Responsive design for mobile and desktop
- Syntax highlighting with GitHub Dark theme
- Smooth animations and transitions

## Development Tips

- Hot Module Replacement (HMR) is enabled for fast development
- TypeScript strict mode is enabled for type safety
- Use the browser console to debug WebSocket messages
- The app reconnects automatically if the connection is lost

## Browser Support

- Modern browsers with ES2020+ support
- WebSocket support required
- Tested on Chrome, Firefox, Safari, Edge

## License

MIT
