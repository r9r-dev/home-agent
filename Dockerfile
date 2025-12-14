# Multi-stage Dockerfile for Home Agent
# Builds both frontend and backend in a single container

# Stage 1: Build frontend
FROM node:20-alpine AS frontend-builder

WORKDIR /app/frontend

# Copy frontend package files
COPY frontend/package*.json ./
RUN npm ci

# Copy frontend source
COPY frontend/ ./

# Build frontend (outputs to ../backend/public)
RUN npm run build

# Stage 2: Build backend
FROM golang:1.21-alpine AS backend-builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY backend/go.* ./
RUN go mod download

# Copy backend source
COPY backend/ ./

# Copy built frontend from previous stage
COPY --from=frontend-builder /app/backend/public ./public

# Build backend binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o home-agent .

# Stage 3: Runtime image
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata curl

# Install Claude Code CLI
# Note: Replace with actual Claude Code installation method
RUN curl -fsSL https://claude.ai/cli/install.sh | sh || \
    echo "Claude CLI installation placeholder - update with actual installation"

# Create non-root user
RUN addgroup -g 1000 homeagent && \
    adduser -D -u 1000 -G homeagent homeagent

# Create directories
RUN mkdir -p /app /data /workspace && \
    chown -R homeagent:homeagent /app /data /workspace

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=backend-builder --chown=homeagent:homeagent /app/home-agent ./
COPY --from=backend-builder --chown=homeagent:homeagent /app/public ./public

# Switch to non-root user
USER homeagent

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

# Set environment variables
ENV PORT=8080 \
    HOST=0.0.0.0 \
    DATABASE_PATH=/data/sessions.db \
    CLAUDE_BIN=/usr/local/bin/claude

# Run the application
CMD ["./home-agent"]
