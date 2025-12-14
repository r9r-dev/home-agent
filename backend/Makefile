.PHONY: help build run clean test dev install

# Default target
help:
	@echo "Home Agent Backend - Make Commands"
	@echo ""
	@echo "Usage:"
	@echo "  make install    - Install dependencies"
	@echo "  make dev        - Run in development mode with hot reload"
	@echo "  make run        - Run the server"
	@echo "  make build      - Build the binary"
	@echo "  make test       - Run tests"
	@echo "  make clean      - Clean build artifacts"
	@echo ""

# Install dependencies
install:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy
	@echo "Dependencies installed"

# Run in development mode
dev:
	@echo "Running in development mode..."
	go run main.go

# Run the server
run: build
	@echo "Running server..."
	./home-agent-server

# Build the binary
build:
	@echo "Building..."
	go build -ldflags="-s -w" -o home-agent-server main.go
	@echo "Build complete: home-agent-server"

# Build for production
build-prod:
	@echo "Building for production..."
	CGO_ENABLED=1 go build -ldflags="-s -w" -o home-agent-server main.go
	@echo "Production build complete"

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f home-agent-server
	rm -f *.db *.db-shm *.db-wal
	rm -rf data/
	@echo "Clean complete"

# Create necessary directories
setup:
	@echo "Setting up directories..."
	mkdir -p data
	mkdir -p public
	@echo "Setup complete"

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	@echo "Linting code..."
	golangci-lint run

# Show project structure
tree:
	@echo "Project structure:"
	tree -I 'vendor|data|public|*.db' -L 3
