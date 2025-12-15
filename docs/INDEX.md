# Documentation

## Guides

| Document | Description |
|----------|-------------|
| [development.md](development.md) | Development setup and workflow |
| [docker.md](docker.md) | Docker deployment and CI/CD |
| [deployment.md](deployment.md) | Production deployment (systemd, nginx) |
| [claude-proxy.md](claude-proxy.md) | Claude Proxy service for host execution |
| [architecture.md](architecture.md) | Technical architecture and design |
| [testing.md](testing.md) | Testing procedures |

## Examples

| File | Description |
|------|-------------|
| [nginx.conf.example](examples/nginx.conf.example) | Nginx reverse proxy configuration |
| [home-agent.service.example](examples/home-agent.service.example) | Systemd service unit |

## Quick Reference

### Development

```bash
./start-dev.sh              # Start dev environment
cd frontend && npm run dev  # Frontend only
cd backend && go run main.go  # Backend only
```

### Docker

```bash
docker pull ghcr.io/r9r-dev/home-agent:latest
docker compose up -d
```

### Testing

```bash
cd frontend && npm run check  # Type check
cd backend && go test ./...   # Go tests
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `ANTHROPIC_API_KEY` | Anthropic API key (required on host for Claude CLI) |
| `PORT` | Server port (default: 8080) |
| `DATABASE_PATH` | SQLite database path |
| `CLAUDE_PROXY_URL` | Claude proxy URL (e.g., `http://192.168.1.100:9090`) |
| `CLAUDE_PROXY_KEY` | API key for proxy authentication |
| `CLAUDE_BIN` | Path to Claude CLI (for local mode only) |

## Claude Execution Modes

Home Agent supports two modes for executing Claude:

1. **Proxy Mode** (recommended for Docker): Connect to a Claude Proxy service running on the host
2. **Local Mode** (for development): Execute Claude CLI directly

See [claude-proxy.md](claude-proxy.md) for setup instructions.
