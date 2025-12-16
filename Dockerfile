# Multi-stage Dockerfile for Home Agent
# Builds frontend and backend in PARALLEL, then combines in runtime stage
# Claude CLI is NOT included - use CLAUDE_PROXY_URL to connect to host proxy

# Stage 1a: Build frontend (runs in parallel with backend)
FROM node:20-alpine AS frontend-builder

# Version from git tag (passed via --build-arg)
ARG APP_VERSION=dev

WORKDIR /app/frontend

# Copy frontend package files
COPY frontend/package*.json ./
RUN --mount=type=cache,target=/root/.npm npm ci

# Copy frontend source
COPY frontend/ ./

# Build frontend with version (outputs to ../backend/public)
RUN APP_VERSION=${APP_VERSION} npm run build

# Stage 1b: Build backend (runs in parallel with frontend)
FROM golang:1.24-alpine AS backend-builder

WORKDIR /app

# Git needed for some Go modules
RUN apk add --no-cache git

# Copy go mod files
COPY backend/go.* ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

# Copy backend source
COPY backend/ ./

# Build backend binary (pure Go SQLite, no CGO needed)
# Note: frontend files are copied directly to runtime stage, not here
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o home-agent .

# Stage 2: Runtime image (minimal Alpine, no Node.js needed)
# This stage waits for both builders to complete, then combines their outputs
FROM alpine:3.19

# Install runtime dependencies only (no sqlite-libs needed with pure Go SQLite)
RUN apk add --no-cache ca-certificates tzdata curl

# Create non-root user
RUN adduser -D -H -u 1000 appuser

# Create directories
RUN mkdir -p /app /data && \
    chown -R appuser:appuser /app /data

# Set working directory
WORKDIR /app

# Copy binary from backend builder
COPY --from=backend-builder --chown=appuser:appuser /app/home-agent ./

# Copy frontend assets directly from frontend builder
COPY --from=frontend-builder --chown=appuser:appuser /app/backend/public ./public

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

# Set environment variables
# CLAUDE_PROXY_URL must be set to point to the host's Claude proxy service
ENV PORT=8080 \
    HOST=0.0.0.0 \
    DATABASE_PATH=/data/sessions.db \
    CLAUDE_PROXY_URL= \
    CLAUDE_PROXY_KEY=

# Run the application
CMD ["./home-agent"]
