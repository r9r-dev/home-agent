#!/bin/bash

# Home Agent Development Startup Script
# This script helps start both backend and frontend in development mode

set -e

echo "🏠 Home Agent - Starting Development Environment"
echo "================================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if CLAUDE_PROXY_URL is set
if [ -z "$CLAUDE_PROXY_URL" ]; then
    echo -e "${YELLOW}Warning: CLAUDE_PROXY_URL not set${NC}"
    echo "Please set the environment variable:"
    echo "  export CLAUDE_PROXY_URL=http://localhost:9090"
    echo ""
    read -p "Continue anyway? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Check if backend dependencies are installed
echo -e "${GREEN}Checking backend dependencies...${NC}"
cd backend
if [ ! -d "vendor" ] && [ ! -f "go.sum" ]; then
    echo "Installing Go dependencies..."
    go mod download
fi

# Build backend
echo -e "${GREEN}Building backend...${NC}"
go build -o home-agent

# Check if frontend dependencies are installed
echo -e "${GREEN}Checking frontend dependencies...${NC}"
cd ../frontend
if [ ! -d "node_modules" ]; then
    echo "Installing npm dependencies..."
    npm install
fi

# Check Node.js version
NODE_VERSION=$(node -v | cut -d'v' -f2 | cut -d'.' -f1)
if [ "$NODE_VERSION" -lt 18 ]; then
    echo -e "${RED}Error: Node.js 18+ required (found v$NODE_VERSION)${NC}"
    exit 1
fi

echo ""
echo -e "${GREEN}✓ All checks passed!${NC}"
echo ""
echo "Starting services..."
echo ""

# Function to cleanup on exit
cleanup() {
    echo ""
    echo "Stopping services..."
    kill $BACKEND_PID 2>/dev/null
    kill $FRONTEND_PID 2>/dev/null
    exit 0
}

trap cleanup SIGINT SIGTERM

# Start backend
cd ../backend
echo -e "${GREEN}Starting backend on http://localhost:8080${NC}"
./home-agent &
BACKEND_PID=$!

# Wait for backend to start
sleep 2

# Start frontend dev server
cd ../frontend
echo -e "${GREEN}Starting frontend dev server on http://localhost:5173${NC}"
npm run dev &
FRONTEND_PID=$!

echo ""
echo -e "${GREEN}✓ Development environment is ready!${NC}"
echo ""
echo "📱 Frontend (Dev):  http://localhost:5173"
echo "🔌 Backend (API):   http://localhost:8080"
echo "🔗 WebSocket:       ws://localhost:8080/ws"
echo ""
echo "Press Ctrl+C to stop all services"
echo ""

# Wait for processes
wait $BACKEND_PID $FRONTEND_PID
