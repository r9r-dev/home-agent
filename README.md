# Home Agent

A web-based chat interface for interacting with Claude Code. Modern frontend built with Svelte and a Go backend with WebSocket support.

## Quick Start

### Prerequisites

- Go 1.21+
- Node.js 18+
- Anthropic API key

### Setup

```bash
# Configure environment
cp .env.example .env
# Edit .env and add your ANTHROPIC_API_KEY

# Start development
./start-dev.sh
```

Open http://localhost:5173 (dev) or http://localhost:8080 (production)

### Docker

```bash
# Pull from GitHub Container Registry
docker pull ghcr.io/r9r-dev/home-agent:latest

# Or build locally
docker compose up -d
```

## Project Structure

```
home-agent/
├── backend/          # Go backend (Fiber + WebSocket)
├── frontend/         # Svelte + TypeScript frontend
├── docs/             # Documentation
└── docker-compose.yml
```

## Features

- Real-time chat with WebSocket streaming
- Markdown rendering with syntax highlighting
- Session persistence with SQLite
- Dark theme optimized for development
- Auto-reconnection with exponential backoff

## Documentation

- [Development Guide](docs/development.md)
- [Docker Deployment](docs/docker.md)
- [Architecture](docs/architecture.md)
- [Testing](docs/testing.md)

## Tech Stack

| Layer | Technology |
|-------|------------|
| Frontend | Svelte 4, TypeScript, Vite |
| Backend | Go 1.21, Fiber, WebSocket |
| Database | SQLite |
| Container | Docker, GitHub Actions |

## License

MIT
