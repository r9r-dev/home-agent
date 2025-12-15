# Claude Proxy Service

The Claude Proxy service allows Home Agent running in Docker to access Claude CLI on the host system, giving Claude full access to the host's resources (bash, network, filesystem, etc.).

## Architecture

```
+-----------------------------------+
|        Docker Container           |
|  +-----------------------------+  |
|  |     Home Agent Backend      |  |
|  |                             |  |
|  |  ProxyClaudeExecutor        |  |
|  |  (WebSocket client)         |  |
|  +-------------+---------------+  |
+----------------|------------------+
                 | WebSocket + API Key
                 | http://<HOST_IP>:9090
                 v
+-----------------------------------+
|           Host System             |
|  +-----------------------------+  |
|  |   Claude Proxy Service      |  |
|  |   (systemd, port 9090)      |  |
|  +-------------+---------------+  |
|                |                  |
|                v                  |
|  +-----------------------------+  |
|  |       Claude CLI            |  |
|  |   (full host access)        |  |
|  +-----------------------------+  |
+-----------------------------------+
```

## Installation

### Quick Install (Recommended)

Install or update Claude Proxy with a single command:

```bash
curl -fsSL https://raw.githubusercontent.com/r9r-dev/home-agent/main/claude-proxy/install.sh | sudo bash
```

The installer will:
1. Download and install the binary to `/opt/claude-proxy/`
2. Generate a secure API key automatically
3. Configure and start the systemd service
4. Display the configuration to use for Home Agent

At the end of installation, you'll see:

```
  ╔═══════════════════════════════════════════════════════════════╗
  ║           Home Agent Configuration                            ║
  ╚═══════════════════════════════════════════════════════════════╝

  Add these environment variables to your Home Agent container:

  ┌─────────────────────────────────────────────────────────────┐
  │  CLAUDE_PROXY_URL=http://192.168.1.100:9090
  │  CLAUDE_PROXY_KEY=<your-generated-key>
  └─────────────────────────────────────────────────────────────┘
```

**Note:** On upgrade, the existing API key is preserved.

### Install Options

Install a specific version:

```bash
curl -fsSL https://raw.githubusercontent.com/r9r-dev/home-agent/main/claude-proxy/install.sh | sudo bash -s -- --version v1.0.0
```

Uninstall:

```bash
curl -fsSL https://raw.githubusercontent.com/r9r-dev/home-agent/main/claude-proxy/install.sh | sudo bash -s -- --uninstall
```

### Manual Install (From Source)

If you prefer to build from source:

```bash
cd claude-proxy

# Install dependencies
go mod tidy

# Option A: Quick install with make
make install

# Option B: Manual installation
./deploy/install.sh
```

### Verify the Proxy is Running

```bash
# Check status
sudo systemctl status claude-proxy

# View logs
sudo journalctl -u claude-proxy -f

# Test health endpoint
curl http://localhost:9090/health
```

### Retrieve API Key (if needed)

If you need to retrieve the API key after installation:

```bash
sudo grep PROXY_API_KEY /etc/systemd/system/claude-proxy.service
```

## Configuration

### Proxy Service (Host)

| Variable | Default | Description |
|----------|---------|-------------|
| `PROXY_PORT` | `9090` | Port to listen on |
| `PROXY_HOST` | `0.0.0.0` | Host to bind to |
| `PROXY_API_KEY` | (empty) | API key for authentication |
| `CLAUDE_BIN` | `claude` | Path to Claude CLI |

### Home Agent (Container)

| Variable | Default | Description |
|----------|---------|-------------|
| `CLAUDE_PROXY_URL` | (empty) | URL of Claude proxy (e.g., `http://192.168.1.100:9090`) |
| `CLAUDE_PROXY_KEY` | (empty) | API key matching proxy |

## Docker Deployment

### Using docker-compose

1. Create `.env` file from example:

```bash
cp .env.example .env
```

2. Edit `.env` and set your host IP:

```bash
HOST_IP=192.168.1.100
CLAUDE_PROXY_KEY=your-secure-api-key
```

3. Start the container:

```bash
docker compose up -d
```

### Manual Docker Run

```bash
docker run -d \
  --name home-agent \
  -p 8080:8080 \
  -v ./data:/data \
  -e CLAUDE_PROXY_URL=http://192.168.1.100:9090 \
  -e CLAUDE_PROXY_KEY=your-secure-api-key \
  home-agent
```

## Security

### API Key Authentication

The proxy uses a simple API key for authentication. All requests must include the `X-API-Key` header.

Generate a secure key:

```bash
openssl rand -hex 32
```

### Network Security

- By default, the proxy binds to `0.0.0.0` to accept connections from Docker
- In production, consider using a firewall to restrict access
- For additional security, use TLS (not included, can be added via reverse proxy)

### Recommended Firewall Rules (Linux)

```bash
# Allow only Docker network to access proxy
sudo ufw allow from 172.16.0.0/12 to any port 9090

# Or restrict to specific IP
sudo ufw allow from 192.168.1.0/24 to any port 9090
```

## Protocol

### WebSocket Messages

**Client -> Proxy (Execute Request):**
```json
{
  "type": "execute",
  "prompt": "Your prompt here",
  "session_id": "optional-session-id"
}
```

**Proxy -> Client (Streaming Response):**
```json
{"type": "chunk", "content": "Partial response..."}
{"type": "session_id", "session_id": "claude-session-id"}
{"type": "done", "content": "Full response", "session_id": "claude-session-id"}
{"type": "error", "error": "Error message"}
```

### HTTP Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check (no auth required) |
| POST | `/api/title` | Generate conversation title |
| GET | `/ws` | WebSocket endpoint for Claude execution |

## Troubleshooting

### Proxy Connection Failed

```
Warning: Claude executor test failed
Make sure the Claude proxy service is running and accessible
```

1. Check if proxy is running: `sudo systemctl status claude-proxy`
2. Check if port is accessible: `curl http://HOST_IP:9090/health`
3. Check firewall rules
4. Verify `CLAUDE_PROXY_URL` is correct

### Authentication Failed

```
proxy returned status 401
```

1. Verify `CLAUDE_PROXY_KEY` matches `PROXY_API_KEY` on the host
2. Check proxy logs: `sudo journalctl -u claude-proxy -f`

### Claude CLI Not Found

```
Warning: Claude binary test failed
```

On the host:
1. Verify Claude CLI is installed: `claude --version`
2. Check `CLAUDE_BIN` path in service file
3. Ensure `ANTHROPIC_API_KEY` is set

## Development

### Run Proxy Locally

```bash
cd claude-proxy
PROXY_PORT=9090 PROXY_HOST=127.0.0.1 go run .
```

### Test with Backend

```bash
cd backend
CLAUDE_PROXY_URL=http://localhost:9090 go run .
```
