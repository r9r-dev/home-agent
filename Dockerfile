# Multi-stage Dockerfile for Home Agent
# Builds both frontend and backend in a single container
# Claude CLI is NOT included - use CLAUDE_PROXY_URL to connect to host proxy

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

# Install build dependencies (including CGO for sqlite)
RUN apk add --no-cache git gcc musl-dev sqlite-dev

# Copy go mod files
COPY backend/go.* ./
RUN go mod download

# Copy backend source
COPY backend/ ./

# Copy built frontend from previous stage
COPY --from=frontend-builder /app/backend/public ./public

# Build backend binary (CGO required for sqlite)
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o home-agent .

# Stage 3: Runtime image (minimal Alpine, no Node.js needed)
FROM alpine:3.19

# Install runtime dependencies only
RUN apk add --no-cache ca-certificates tzdata curl sqlite-libs

# Create non-root user
RUN adduser -D -H -u 1000 appuser

# Create directories
RUN mkdir -p /app /data && \
    chown -R appuser:appuser /app /data

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=backend-builder --chown=appuser:appuser /app/home-agent ./
COPY --from=backend-builder --chown=appuser:appuser /app/public ./public

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
