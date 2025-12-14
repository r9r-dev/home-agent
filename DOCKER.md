# Docker Deployment Guide

Guide for deploying Home Agent using Docker and Docker Compose.

## Prerequisites

- Docker 20.10+
- Docker Compose 2.0+
- Anthropic API key

## Quick Start

### 1. Setup Environment

Create `.env` file:
```bash
cp .env.example .env
```

Edit `.env` and add your API key:
```env
ANTHROPIC_API_KEY=your_api_key_here
PORT=8080
HOST=0.0.0.0
DATABASE_PATH=/data/sessions.db
```

### 2. Build and Run

Using Make (recommended):
```bash
make build
make run
```

Or using Docker Compose directly:
```bash
docker compose build
docker compose up -d
```

### 3. Access Application

Open browser to: http://localhost:8080

## Makefile Commands

```bash
make help      # Show available commands
make build     # Build Docker image
make run       # Start containers
make stop      # Stop containers
make restart   # Restart containers
make logs      # View logs
make shell     # Open shell in container
make clean     # Remove containers and images
make dev       # Run development environment (no Docker)
make test      # Run tests
```

## Docker Compose Configuration

The `docker-compose.yml` file defines:

- **Service**: home-agent
- **Port**: 8080 (configurable)
- **Volumes**:
  - `./data` - Persistent database storage
  - `./workspace` - Optional workspace for Claude Code
- **Network**: home-agent-network (bridge)
- **Health Check**: HTTP endpoint at /health

## Dockerfile Overview

Multi-stage build process:

1. **Frontend Build Stage**:
   - Node.js 20 Alpine
   - Builds Svelte + TypeScript frontend
   - Outputs to `backend/public`

2. **Backend Build Stage**:
   - Go 1.21 Alpine
   - Compiles Go backend
   - Includes built frontend

3. **Runtime Stage**:
   - Alpine Linux (minimal)
   - Non-root user (UID 1000)
   - Health checks enabled
   - Claude CLI installed

## Configuration

### Environment Variables

Set in `.env` file:

```env
# Required
ANTHROPIC_API_KEY=sk-ant-your-key-here

# Optional (with defaults)
PORT=8080
HOST=0.0.0.0
DATABASE_PATH=/data/sessions.db
CLAUDE_BIN=/usr/local/bin/claude
```

### Volumes

**Data Volume** (`./data`):
- Contains SQLite database
- Persists conversation history
- Backup regularly

**Workspace Volume** (`./workspace`):
- Optional workspace for Claude Code
- Can be used for file operations
- Shared between host and container

### Ports

Default port is 8080. To change:

In `.env`:
```env
PORT=3000
```

In `docker-compose.yml`:
```yaml
ports:
  - "3000:3000"
```

## Development vs Production

### Development (Without Docker)

```bash
# Use development scripts
./start-dev.sh

# Or manually
cd backend && go run main.go
cd frontend && npm run dev
```

### Production (With Docker)

```bash
# Build and run
make build
make run

# Or with docker-compose
docker compose up -d
```

## Monitoring

### View Logs

```bash
# All logs
make logs

# Or with docker-compose
docker compose logs -f

# Specific service
docker compose logs -f home-agent
```

### Health Check

```bash
# Manual check
curl http://localhost:8080/health

# Docker health status
docker ps
```

### Access Container

```bash
# Open shell
make shell

# Or with docker-compose
docker compose exec home-agent /bin/sh
```

## Troubleshooting

### Container Won't Start

```bash
# Check logs
docker compose logs home-agent

# Check container status
docker ps -a

# Verify environment variables
docker compose config
```

### Database Issues

```bash
# Check data volume
ls -la ./data

# Permissions
chmod 755 ./data
```

### Port Already in Use

```bash
# Change port in .env
PORT=3000

# Update docker-compose.yml
ports:
  - "3000:3000"

# Restart
make restart
```

### Build Fails

```bash
# Clean and rebuild
make clean
make build

# Check Docker version
docker --version
docker compose version
```

## Backup and Restore

### Backup Database

```bash
# Create backup
docker compose exec home-agent cp /data/sessions.db /data/backup-$(date +%Y%m%d).db

# Copy to host
docker compose cp home-agent:/data/backup-*.db ./backups/
```

### Restore Database

```bash
# Stop container
docker compose stop

# Copy backup to data volume
cp ./backups/sessions.db ./data/

# Start container
docker compose start
```

## Updating

### Update Application

```bash
# Pull latest code
git pull

# Rebuild image
make build

# Restart containers
make restart
```

### Update Dependencies

Frontend:
```bash
cd frontend
npm update
```

Backend:
```bash
cd backend
go get -u ./...
go mod tidy
```

Then rebuild:
```bash
make build
make restart
```

## Security

### Best Practices

1. **API Keys**: Never commit `.env` to git
2. **User**: Container runs as non-root (UID 1000)
3. **Network**: Use isolated Docker network
4. **Updates**: Keep base images updated

### Scanning

```bash
# Scan image for vulnerabilities
docker scan home-agent:latest

# Or use trivy
trivy image home-agent:latest
```

## Performance

### Resource Limits

Add to `docker-compose.yml`:

```yaml
services:
  home-agent:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G
        reservations:
          cpus: '1'
          memory: 512M
```

### Optimization

1. **Multi-stage builds**: Already implemented
2. **Layer caching**: Optimized Dockerfile order
3. **Minimal base image**: Alpine Linux
4. **No unnecessary files**: `.dockerignore` configured

## Production Deployment

### With Reverse Proxy

Use Nginx/Traefik in front of container:

```yaml
services:
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
      - ./certs:/etc/nginx/certs
    depends_on:
      - home-agent
```

### With SSL

Use Let's Encrypt with certbot:

```yaml
services:
  certbot:
    image: certbot/certbot
    volumes:
      - ./certs:/etc/letsencrypt
```

### Scaling

For multiple instances:

```bash
docker compose up -d --scale home-agent=3
```

Add load balancer (Nginx/HAProxy) in front.

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Build and Push

on:
  push:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Build image
        run: docker build -t home-agent .
      - name: Push to registry
        run: docker push home-agent
```

## Docker Hub

### Build for Multiple Platforms

```bash
# Setup buildx
docker buildx create --use

# Build for multiple architectures
docker buildx build --platform linux/amd64,linux/arm64 -t home-agent:latest .
```

### Push to Registry

```bash
# Tag image
docker tag home-agent:latest yourusername/home-agent:latest

# Push
docker push yourusername/home-agent:latest
```

## Common Commands

```bash
# View running containers
docker ps

# View all containers
docker ps -a

# View images
docker images

# Remove stopped containers
docker container prune

# Remove unused images
docker image prune

# Remove all unused data
docker system prune -a

# View resource usage
docker stats

# Export image
docker save home-agent:latest | gzip > home-agent.tar.gz

# Import image
docker load < home-agent.tar.gz
```

## Support

For Docker-related issues:

1. Check logs: `make logs`
2. Verify configuration: `docker compose config`
3. Test health: `curl http://localhost:8080/health`
4. Review Dockerfile and docker-compose.yml
5. Check Docker daemon: `docker info`

## Resources

- [Docker Documentation](https://docs.docker.com/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Docker Best Practices](https://docs.docker.com/develop/dev-best-practices/)
- [Alpine Linux](https://alpinelinux.org/)

---

For non-Docker deployment, see [DEPLOYMENT.md](DEPLOYMENT.md).
